package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/qa-guru/selenoid/config"
	"github.com/qa-guru/selenoid/session"
	"github.com/stretchr/testify/require"
)

type stubPool struct {
	id, wdURL string
	err       error
	reserved  int
	released  []string
}

func (s *stubPool) Reserve(protocol, browser, owner string, loopback bool) (string, string, error) {
	s.reserved++
	if s.err != nil {
		return "", "", s.err
	}
	return s.id, s.wdURL, nil
}

func (s *stubPool) Release(slotID, owner string) error {
	s.released = append(s.released, slotID)
	return nil
}

type stubStarter struct {
	err error
	ss  *StartedService
}

func (s *stubStarter) StartWithCancel() (*StartedService, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.ss != nil {
		return s.ss, nil
	}
	u, _ := url.Parse("http://127.0.0.1:9/")
	return &StartedService{Url: u, Cancel: func() {}}, nil
}

func chromeDriverConfig() *config.Config {
	conf := config.NewConfig()
	conf.Browsers["chrome"] = config.Versions{
		Default: "149.0",
		Versions: map[string]*config.Browser{
			"149.0": {
				Image: []interface{}{"chromedriver"},
				Path:  "/",
			},
		},
	}
	conf.Browsers["firefox"] = config.Versions{
		Default: "151.0",
		Versions: map[string]*config.Browser{
			"151.0": {
				Image: []interface{}{"geckodriver"},
				Path:  "/",
			},
		},
	}
	return conf
}

func TestWarmEligible(t *testing.T) {
	require.True(t, warmEligible("chrome", "152.0", "152.0", session.Caps{}))
	require.True(t, warmEligible("Chrome", "152.0", "152.0", session.Caps{}))
	require.True(t, warmEligible("chrome", "", "152.0", session.Caps{}))
	require.False(t, warmEligible("firefox", "152.0", "152.0", session.Caps{}))
	require.False(t, warmEligible("chrome", "152.0", "152.0", session.Caps{Video: true}))
	require.False(t, warmEligible("chrome", "152.0", "152.0", session.Caps{VNC: true}))
	require.False(t, warmEligible("chrome", "152.0", "152.0", session.Caps{HAR: true}))
	require.False(t, warmEligible("chrome", "148.0", "152.0", session.Caps{}))
	require.False(t, warmEligible("chrome", "148", "152.0", session.Caps{}))
	require.False(t, warmEligible("chrome", "151.0", "152.0", session.Caps{}))
	require.False(t, warmEligible("chrome", "152.0-min", "152.0", session.Caps{}))
	require.True(t, warmEligible("chrome", "152.0", "152.0", session.Caps{}))
}

func TestChromeMajor(t *testing.T) {
	require.Equal(t, "148", chromeMajor("148"))
	require.Equal(t, "148", chromeMajor("148.0"))
	require.Equal(t, "148", chromeMajor("148.0-min"))
	require.Equal(t, "152", chromeMajor("152.0"))
	require.Equal(t, "", chromeMajor(""))
	require.Equal(t, "", chromeMajor("latest"))
}

func TestWrapWarmAttachesChrome(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{}, cold, pool, &Environment{StartupTimeout: time.Second})
	fb, ok := got.(*fallbackStarter)
	require.True(t, ok)
	require.Equal(t, 1, pool.reserved)
	att, ok := fb.primary.(*WarmAttach)
	require.True(t, ok)
	require.Equal(t, "pool-chrome-1", att.SlotID)
	tagged, ok := fb.fallback.(*poolTagged)
	require.True(t, ok)
	require.Equal(t, cold, tagged.starter)
}

// taggedCold unwraps the cold-side poolTagged routing wraps around.
func taggedCold(t *testing.T, got Starter, cold *stubStarter) *poolTagged {
	t.Helper()
	tagged, ok := got.(*poolTagged)
	require.True(t, ok, "expected *poolTagged, got %T", got)
	require.Equal(t, cold, tagged.starter)
	require.Equal(t, "cold", tagged.pool)
	return tagged
}

func TestWrapWarmSkipsCompatVersion(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "148.0", "152.0", session.Caps{}, cold, pool, nil)
	require.Equal(t, "version-pin", taggedCold(t, got, cold).note)
	require.Equal(t, 0, pool.reserved)
}

func TestWrapWarmSkipsWhenNoLoopbackURL(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://warm-chrome-1:4444/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{}, cold, pool, nil)
	require.Equal(t, "warm-unavailable", taggedCold(t, got, cold).note)
	require.Equal(t, []string{"pool-chrome-1"}, pool.released)
}

func TestWrapWarmSkipsOnReserveError(t *testing.T) {
	pool := &stubPool{err: errors.New("409")}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{}, cold, pool, nil)
	require.Equal(t, "warm-unavailable", taggedCold(t, got, cold).note)
}

func TestWrapWarmSkipsVideoCaps(t *testing.T) {
	pool := &stubPool{id: "x", wdURL: "http://127.0.0.1:1/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Video: true}, cold, pool, nil)
	require.Equal(t, "sidecar-caps", taggedCold(t, got, cold).note)
	require.Equal(t, 0, pool.reserved)
}

func TestWrapWarmColdSelectorSkipsReserve(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Pool: "cold"}, cold, pool, nil)
	tagged := taggedCold(t, got, cold)
	require.Equal(t, "", tagged.note)
	require.Equal(t, 0, pool.reserved)
}

func TestWrapWarmColdSelectorViaLabels(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	caps := session.Caps{Labels: map[string]string{"pool": "cold"}}
	got := wrapWarm(7, "chrome", "152.0", "152.0", caps, cold, pool, nil)
	taggedCold(t, got, cold)
	require.Equal(t, 0, pool.reserved)
}

func TestWrapWarmStrictWarmAttaches(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Pool: "warm"}, cold, pool, nil)
	att, ok := got.(*WarmAttach)
	require.True(t, ok, "strict warm must attach directly, got %T", got)
	require.Equal(t, "pool-chrome-1", att.SlotID)
}

func TestWrapWarmStrictWarmRefusesSidecar(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Pool: "warm", HAR: true}, cold, pool, nil)
	fail, ok := got.(*failStarter)
	require.True(t, ok, "expected *failStarter, got %T", got)
	require.Contains(t, fail.err.Error(), "forcedPoolReason=sidecar-caps")
	require.Contains(t, fail.err.Error(), "enableHAR")
	require.Equal(t, 0, pool.reserved)
}

func TestWrapWarmStrictWarmRefusesWhenBusy(t *testing.T) {
	pool := &stubPool{err: errors.New("409")}
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Pool: "warm"}, cold, pool, nil)
	fail, ok := got.(*failStarter)
	require.True(t, ok, "expected *failStarter, got %T", got)
	require.Contains(t, fail.err.Error(), "forcedPoolReason=warm-unavailable")
}

func TestWrapWarmStrictWarmNoPool(t *testing.T) {
	cold := &stubStarter{}
	got := wrapWarm(7, "chrome", "152.0", "152.0", session.Caps{Pool: "warm"}, cold, nil, nil)
	fail, ok := got.(*failStarter)
	require.True(t, ok, "expected *failStarter, got %T", got)
	require.Contains(t, fail.err.Error(), "forcedPoolReason=no-warm-pool")
}

func TestFailStarterReturnsError(t *testing.T) {
	f := &failStarter{err: errors.New("nope")}
	_, err := f.StartWithCancel()
	require.EqualError(t, err, "nope")
}

func TestPoolTaggedStampsServedPool(t *testing.T) {
	u, _ := url.Parse("http://127.0.0.1:4444/")
	tagged := &poolTagged{
		starter: &stubStarter{ss: &StartedService{Url: u, Cancel: func() {}}},
		pool:    "cold",
		note:    "sidecar-caps",
	}
	ss, err := tagged.StartWithCancel()
	require.NoError(t, err)
	require.Equal(t, "cold", ss.ServedPool)
	require.Equal(t, "sidecar-caps", ss.PoolNote)
}

func TestFallbackStarterUsesColdWhenWarmFails(t *testing.T) {
	coldURL, _ := url.Parse("http://127.0.0.1:4444/")
	f := &fallbackStarter{
		requestId: 1,
		primary:   &stubStarter{err: errors.New("wait timeout")},
		fallback:  &stubStarter{ss: &StartedService{Url: coldURL, Cancel: func() {}}},
	}
	ss, err := f.StartWithCancel()
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:4444", ss.Url.Host)
}

func TestWarmAttachStartAndRelease(t *testing.T) {
	drv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer drv.Close()

	pool := &stubPool{}
	att := &WarmAttach{
		RequestId:    3,
		SlotID:       "pool-chrome-1",
		WebdriverURL: drv.URL + "/",
		Timeout:      time.Second,
		Pool:         pool,
	}
	ss, err := att.StartWithCancel()
	require.NoError(t, err)
	want, err := url.Parse(drv.URL)
	require.NoError(t, err)
	require.Equal(t, want.Host, ss.Url.Host)
	ss.Cancel()
	require.Equal(t, []string{"pool-chrome-1"}, pool.released)
}

func TestWarmAttachWaitFailReleases(t *testing.T) {
	pool := &stubPool{}
	att := &WarmAttach{
		RequestId:    4,
		SlotID:       "pool-chrome-1",
		WebdriverURL: "http://127.0.0.1:1/",
		Timeout:      50 * time.Millisecond,
		Pool:         pool,
	}
	_, err := att.StartWithCancel()
	require.Error(t, err)
	require.Equal(t, []string{"pool-chrome-1"}, pool.released)
}

func TestFindWrapsWarmForChromeDriver(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	m := &DefaultManager{
		Environment: &Environment{StartupTimeout: time.Second},
		Config:      chromeDriverConfig(),
		WarmPool:    pool,
	}
	starter, ok := m.Find(session.Caps{Name: "chrome"}, 9)
	require.True(t, ok)
	_, isFB := starter.(*fallbackStarter)
	require.True(t, isFB)
	require.Equal(t, 1, pool.reserved)
}

func TestFindFirefoxStaysCold(t *testing.T) {
	pool := &stubPool{id: "pool-chrome-1", wdURL: "http://127.0.0.1:14441/"}
	m := &DefaultManager{
		Environment: &Environment{},
		Config:      chromeDriverConfig(),
		WarmPool:    pool,
	}
	starter, ok := m.Find(session.Caps{Name: "firefox"}, 9)
	require.True(t, ok)
	tagged, isTagged := starter.(*poolTagged)
	require.True(t, isTagged)
	_, isDriver := tagged.starter.(*Driver)
	require.True(t, isDriver)
	require.Equal(t, "non-chrome", tagged.note)
	require.Equal(t, 0, pool.reserved)
}
