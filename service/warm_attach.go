package service

import (
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
	Release(slotID string) error
}

// WarmAttach proxies to a reserved warm ChromeDriver. Cancel releases the slot.
type WarmAttach struct {
	RequestId    uint64
	SlotID       string
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
		Url:      u,
		HostPort: session.HostPort{Selenium: u.Host},
		Origin:   u.Host,
		Cancel:   w.release,
	}, nil
}

func (w *WarmAttach) release() {
	if w.Pool == nil || w.SlotID == "" {
		return
	}
	s := time.Now()
	err := w.Pool.Release(w.SlotID)
	if err != nil {
		log.Printf("[%d] [WARM_POOL_RELEASE_FAILED] [%s] [%v]", w.RequestId, w.SlotID, err)
		return
	}
	log.Printf("[%d] [WARM_POOL_RELEASED] [%s] [%.2fs]", w.RequestId, w.SlotID, info.SecondsSince(s))
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
	return f.fallback.StartWithCancel()
}

func warmEligible(browserName string, caps session.Caps) bool {
	if !strings.EqualFold(browserName, "chrome") {
		return false
	}
	if caps.Video || caps.VNC || caps.HAR {
		return false
	}
	return true
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
			_ = pool.Release(id)
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
		WebdriverURL: wdURL,
		Timeout:      timeout,
		Pool:         pool,
	}
}

func wrapWarm(requestId uint64, browserName string, caps session.Caps, cold Starter, pool WarmPool, env *Environment) Starter {
	if cold == nil || pool == nil || !warmEligible(browserName, caps) {
		return cold
	}
	att := tryWarmAttach(pool, env, requestId)
	if att == nil {
		log.Printf("[%d] [WARM_POOL_FALLBACK_COLD] [no loopback chrome slot]", requestId)
		return cold
	}
	return &fallbackStarter{requestId: requestId, primary: att, fallback: cold}
}
