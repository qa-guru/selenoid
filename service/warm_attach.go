package service

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/qa-guru/selenoid/info"
	"github.com/qa-guru/selenoid/session"
	"github.com/qa-guru/selenoid/warm"
)

const warmAttachWait = 2 * time.Second

// WarmPool reserves/releases pre-started browser slots (implemented by warm.Client).
type WarmPool interface {
	Reserve(protocol, browser, owner string, loopback bool) (slotID, webdriverURL string, err error)
	Release(slotID, owner string) error
}

// WarmAttach proxies to a reserved warm ChromeDriver. Cancel releases the slot.
type WarmAttach struct {
	RequestId    uint64
	SlotID       string
	Owner        string
	WebdriverURL string
	Timeout      time.Duration
	Pool         WarmPool
}

func (w *WarmAttach) StartWithCancel() (*StartedService, error) {
	u, err := url.Parse(strings.TrimSpace(w.WebdriverURL))
	if err != nil || u.Host == "" {
		w.release()
		return nil, fmt.Errorf("warm slot %s: bad webdriver url %q", w.SlotID, w.WebdriverURL)
	}
	timeout := w.Timeout
	if timeout <= 0 {
		timeout = warmAttachWait
	}
	if err := wait(u.String(), timeout); err != nil {
		w.release()
		return nil, err
	}
	log.Printf("[%d] [USING_WARM_POOL] [%s] [%s]", w.RequestId, w.SlotID, u.String())
	log.Printf("[%d] [PROXY_TO] [%s]", w.RequestId, u.String())
	return &StartedService{
		Url:        u,
		HostPort:   session.HostPort{Selenium: u.Host},
		Origin:     u.Host,
		Cancel:     w.release,
		ServedPool: "warm",
		ServedSlot: w.SlotID,
	}, nil
}

func (w *WarmAttach) release() {
	if w.Pool == nil || w.SlotID == "" {
		return
	}
	s := time.Now()
	err := w.Pool.Release(w.SlotID, w.Owner)
	if err != nil {
		log.Printf("[%d] [WARM_POOL_RELEASE_FAILED] [%s] [%v]", w.RequestId, w.SlotID, err)
		return
	}
	log.Printf("[%d] [WARM_POOL_RELEASED] [%s] [%.2fs]", w.RequestId, w.SlotID, info.SecondsSince(s))
}

// failStarter fails StartWithCancel with a fixed error — used for strict
// pool=warm requests that cannot be served, so the client sees the reason
// instead of a silent cold fallback.
type failStarter struct {
	err error
}

func (f *failStarter) StartWithCancel() (*StartedService, error) {
	return nil, f.err
}

// poolTagged wraps a cold starter and stamps the served pool (+ optional
// forcedPoolReason token) on the StartedService for response telemetry.
type poolTagged struct {
	starter Starter
	pool    string
	note    string
}

func (t *poolTagged) StartWithCancel() (*StartedService, error) {
	ss, err := t.starter.StartWithCancel()
	if err != nil || ss == nil {
		return ss, err
	}
	ss.ServedPool = t.pool
	if ss.PoolNote == "" {
		ss.PoolNote = t.note
	}
	return ss, nil
}

type fallbackStarter struct {
	requestId uint64
	primary   Starter
	fallback  Starter
}

func (f *fallbackStarter) StartWithCancel() (*StartedService, error) {
	ss, err := f.primary.StartWithCancel()
	if err == nil {
		return ss, nil
	}
	log.Printf("[%d] [WARM_POOL_FALLBACK_COLD] [%v]", f.requestId, err)
	ss, err = f.fallback.StartWithCancel()
	if err == nil && ss != nil {
		ss.ServedPool = "cold"
		if ss.PoolNote == "" {
			ss.PoolNote = "warm-attach-failed"
		}
	}
	return ss, err
}

func chromeMajor(version string) string {
	v := strings.TrimSpace(version)
	if v == "" || strings.EqualFold(v, "latest") {
		return ""
	}
	v = strings.TrimSuffix(v, "-min")
	if i := strings.IndexByte(v, '.'); i > 0 {
		v = v[:i]
	}
	return v
}

// warmIneligible returns the detail text and short token explaining why a
// warm slot cannot serve this request ("", "" when eligible). The token is
// the forcedPoolReason contract surfaced to clients.
func warmIneligible(browserName, resolvedVersion, catalogDefault string, caps session.Caps) (detail, token string) {
	if !strings.EqualFold(browserName, "chrome") {
		return fmt.Sprintf("warm pool serves chrome only, requested %q", browserName), "non-chrome"
	}
	var sidecars []string
	if caps.Video {
		sidecars = append(sidecars, "enableVideo")
	}
	if caps.VNC {
		sidecars = append(sidecars, "enableVNC")
	}
	if caps.HAR {
		sidecars = append(sidecars, "enableHAR")
	}
	if len(sidecars) > 0 {
		return fmt.Sprintf("%s require a cold container (sidecars are baked at container start)", strings.Join(sidecars, "+")), "sidecar-caps"
	}
	// Warm WD slots are headed catalog-default Chrome. Compat pins (148, 151)
	// and *-min tags must cold-start their catalog image so students get that version.
	if strings.HasSuffix(strings.TrimSpace(resolvedVersion), "-min") {
		return fmt.Sprintf("version %q is a -min image, not served by warm slots", resolvedVersion), "min-image"
	}
	req := chromeMajor(resolvedVersion)
	def := chromeMajor(catalogDefault)
	if req != "" && def != "" && req != def {
		return fmt.Sprintf("version %q pins a different Chrome major than the warm slot default %q", resolvedVersion, catalogDefault), "version-pin"
	}
	return "", ""
}

func warmEligible(browserName, resolvedVersion, catalogDefault string, caps session.Caps) bool {
	detail, _ := warmIneligible(browserName, resolvedVersion, catalogDefault, caps)
	return detail == ""
}

func tryWarmAttach(pool WarmPool, env *Environment, requestId uint64) Starter {
	if pool == nil {
		return nil
	}
	owner := fmt.Sprintf("hub-%d", requestId)
	id, wdURL, err := pool.Reserve("webdriver", "chrome", owner, true)
	if err != nil {
		return nil
	}
	if id == "" || !warm.IsLoopbackURL(wdURL) {
		if id != "" {
			_ = pool.Release(id, owner)
		}
		return nil
	}
	timeout := warmAttachWait
	if env != nil && env.StartupTimeout > 0 && env.StartupTimeout < timeout {
		timeout = env.StartupTimeout
	}
	return &WarmAttach{
		RequestId:    requestId,
		SlotID:       id,
		Owner:        owner,
		WebdriverURL: wdURL,
		Timeout:      timeout,
		Pool:         pool,
	}
}

// wrapWarm routes the starter by the explicit pool selector
// (selenoid:options.pool or labels.pool): unset/"cold" is the hub default and
// always cold-starts; "warm" is strict — a served warm slot or a visible
// error, never a silent cold; "auto" keeps prefer-warm behavior but tags the
// actual served pool and a forcedPoolReason token so the client sees the
// fallback in the response.
func wrapWarm(requestId uint64, browserName, resolvedVersion, catalogDefault string, caps session.Caps, cold Starter, pool WarmPool, env *Environment) Starter {
	if cold == nil {
		return nil
	}
	requested := caps.RequestedPool()
	if requested == "" || requested == "cold" {
		return &poolTagged{starter: cold, pool: "cold"}
	}
	if pool == nil {
		if requested == "warm" {
			return &failStarter{err: errors.New("pool=warm requested but the hub has no warm pool configured (forcedPoolReason=no-warm-pool)")}
		}
		return &poolTagged{starter: cold, pool: "cold", note: "no-warm-pool"}
	}
	detail, token := warmIneligible(browserName, resolvedVersion, catalogDefault, caps)
	if detail != "" {
		if requested == "warm" {
			return &failStarter{err: fmt.Errorf("pool=warm unavailable: %s (forcedPoolReason=%s)", detail, token)}
		}
		return &poolTagged{starter: cold, pool: "cold", note: token}
	}
	att := tryWarmAttach(pool, env, requestId)
	if att == nil {
		if requested == "warm" {
			return &failStarter{err: errors.New("pool=warm requested but no warm slot is free (forcedPoolReason=warm-unavailable)")}
		}
		log.Printf("[%d] [WARM_POOL_FALLBACK_COLD] [no loopback chrome slot]", requestId)
		return &poolTagged{starter: cold, pool: "cold", note: "warm-unavailable"}
	}
	if requested == "warm" {
		// Strict: a served warm slot or a visible startup error — never cold.
		return att
	}
	return &fallbackStarter{
		requestId: requestId,
		primary:   att,
		fallback:  &poolTagged{starter: cold, pool: "cold", note: "warm-attach-failed"},
	}
}
