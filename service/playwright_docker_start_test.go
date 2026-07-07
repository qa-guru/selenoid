package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aerokube/selenoid/session"
	ctr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const playwrightDockerAPIVersion = "1.45"

const (
	mockBrowserContainerID = "pw-browser-001"
	mockVideoContainerID   = "pw-video-001"
)

type playwrightDockerMock struct {
	server         *httptest.Server
	cli            *client.Client
	playwrightPort string

	mu sync.Mutex

	createCalls      int
	started          map[string]int
	inspectCalls     int
	removed          []string
	killed           []string
	waitCalls        int

	failCreate       bool
	failStartBrowser bool
	failStartVideo   bool
	failVideoCreate  bool
	emptyBindings    bool
}

func newPlaywrightDockerMockWithPort(t *testing.T, playwrightHostPort string) *playwrightDockerMock {
	t.Helper()

	m := &playwrightDockerMock{
		started:        make(map[string]int),
		playwrightPort: playwrightHostPort,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.serveHTTP))

	t.Setenv("DOCKER_HOST", "tcp://"+m.server.Listener.Addr().String())
	t.Setenv("DOCKER_API_VERSION", playwrightDockerAPIVersion)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })
	m.cli = cli

	return m
}

func newPlaywrightDockerMock(t *testing.T, playwrightSrv *httptest.Server) *playwrightDockerMock {
	t.Helper()
	return newPlaywrightDockerMockWithPort(t, strings.TrimPrefix(playwrightSrv.URL, "http://"))
}

func (m *playwrightDockerMock) serveHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := "/v" + playwrightDockerAPIVersion
	path := r.URL.Path

	switch {
	case r.Method == http.MethodPost && path == prefix+"/containers/create":
		m.createCalls++
		if m.failCreate {
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		if m.createCalls > 1 && m.failVideoCreate {
			http.Error(w, "video create failed", http.StatusInternalServerError)
			return
		}
		id := mockBrowserContainerID
		if m.createCalls > 1 {
			id = mockVideoContainerID
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"Id":"%s","Warnings":[]}`, id)

	case r.Method == http.MethodPost && strings.HasSuffix(path, "/start"):
		id := containerIDFromPath(path, "/start")
		if id == mockBrowserContainerID && m.failStartBrowser {
			http.Error(w, "start failed", http.StatusInternalServerError)
			return
		}
		if id == mockVideoContainerID && m.failStartVideo {
			http.Error(w, "video start failed", http.StatusInternalServerError)
			return
		}
		m.started[id]++
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodGet && strings.HasSuffix(path, "/json"):
		m.inspectCalls++
		hostPort := m.playwrightPort
		if m.emptyBindings {
			hostPort = ""
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, inspectJSON(hostPort), hostPort)

	case r.Method == http.MethodPost && strings.HasSuffix(path, "/kill"):
		id := containerIDFromPath(path, "/kill")
		m.killed = append(m.killed, id)
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodPost && strings.HasSuffix(path, "/wait"):
		m.waitCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"StatusCode":0}`))

	case r.Method == http.MethodDelete && strings.HasPrefix(path, prefix+"/containers/"):
		id := strings.TrimPrefix(path, prefix+"/containers/")
		m.removed = append(m.removed, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.NotFound(w, r)
	}
}

func containerIDFromPath(path, suffix string) string {
	trimmed := strings.TrimSuffix(path, suffix)
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func inspectJSON(hostPort string) string {
	binding := ""
	if hostPort != "" {
		p := hostPort
		if idx := strings.LastIndex(hostPort, ":"); idx >= 0 {
			p = hostPort[idx+1:]
		}
		binding = fmt.Sprintf(`"3000/tcp":[{"HostIp":"0.0.0.0","HostPort":"%s"}]`, p)
	} else {
		binding = `"3000/tcp":[]`
	}
	return `{
		"Id":"` + mockBrowserContainerID + `",
		"NetworkSettings":{
			"Ports":{` + binding + `},
			"Networks":{
				"bridge":{"IPAddress":"172.17.0.2"}
			}
		}
	}`
}

func testPlaywrightEnvironment() Environment {
	dir, err := os.MkdirTemp("", "pw-video-out")
	if err != nil {
		panic(err)
	}
	return Environment{
		InDocker:             false,
		Network:              DefaultContainerNetwork,
		StartupTimeout:       2 * time.Second,
		SessionDeleteTimeout: 500 * time.Millisecond,
		VideoContainerImage:  "qaguru/video-recorder:latest",
		VideoOutputDir:       dir,
	}
}

func newPlaywrightDockerStarter(t *testing.T, mock *playwrightDockerMock, caps session.Caps, env Environment) *PlaywrightDocker {
	t.Helper()
	browser := configBrowser("3000")
	if browser.Path == "" {
		browser.Path = "/"
	}
	return &PlaywrightDocker{
		ServiceBase: ServiceBase{RequestId: 42, Service: browser},
		Environment: env,
		Caps:        caps,
		LogConfig:   &ctr.LogConfig{},
		Client:      mock.cli,
	}
}

func newPlaywrightHTTPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestPlaywrightDockerStartWithCancel_Success(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	defer mock.server.Close()

	caps := session.Caps{Headless: true, ScreenResolution: "1280x720"}
	started, err := newPlaywrightDockerStarter(t, mock, caps, testPlaywrightEnvironment()).StartWithCancel()
	require.NoError(t, err)
	require.NotNil(t, started)

	assert.Equal(t, "ws", started.Url.Scheme)
	assert.Equal(t, "/", started.Url.Path)
	assert.Contains(t, started.Url.Host, "127.0.0.1")
	assert.Equal(t, mockBrowserContainerID, started.Container.ID)
	assert.NotEmpty(t, started.HostPort.Playwright)

	mock.mu.Lock()
	assert.Equal(t, 1, mock.createCalls)
	assert.Equal(t, 1, mock.started[mockBrowserContainerID])
	assert.GreaterOrEqual(t, mock.inspectCalls, 1)
	mock.mu.Unlock()

	started.Cancel()
	mock.mu.Lock()
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_CreateError(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	mock.failCreate = true
	defer mock.server.Close()

	_, err := newPlaywrightDockerStarter(t, mock, session.Caps{Headless: true}, testPlaywrightEnvironment()).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create playwright container")

	mock.mu.Lock()
	assert.Equal(t, 1, mock.createCalls)
	assert.Empty(t, mock.removed)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_StartError(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	mock.failStartBrowser = true
	defer mock.server.Close()

	_, err := newPlaywrightDockerStarter(t, mock, session.Caps{Headless: true}, testPlaywrightEnvironment()).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start playwright container")

	mock.mu.Lock()
	assert.Equal(t, 1, mock.createCalls)
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_WaitInspectError(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	mock.emptyBindings = true
	defer mock.server.Close()

	env := testPlaywrightEnvironment()
	env.StartupTimeout = 150 * time.Millisecond

	_, err := newPlaywrightDockerStarter(t, mock, session.Caps{Headless: true}, env).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no bindings available")

	mock.mu.Lock()
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_WaitServerError(t *testing.T) {
	// No playwright HTTP server — port 1 is closed.
	mock := newPlaywrightDockerMockWithPort(t, "127.0.0.1:1")
	defer mock.server.Close()

	env := testPlaywrightEnvironment()
	env.StartupTimeout = 200 * time.Millisecond

	_, err := newPlaywrightDockerStarter(t, mock, session.Caps{Headless: true}, env).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wait playwright server")

	mock.mu.Lock()
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_Video(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	defer mock.server.Close()

	caps := session.Caps{
		Headless:         true,
		Video:            true,
		VideoName:        "test.mp4",
		ScreenResolution: "1280x720",
	}
	started, err := newPlaywrightDockerStarter(t, mock, caps, testPlaywrightEnvironment()).StartWithCancel()
	require.NoError(t, err)
	require.NotNil(t, started.Cancel)

	mock.mu.Lock()
	assert.Equal(t, 2, mock.createCalls)
	assert.Equal(t, 1, mock.started[mockBrowserContainerID])
	assert.Equal(t, 1, mock.started[mockVideoContainerID])
	mock.mu.Unlock()

	started.Cancel()

	mock.mu.Lock()
	assert.Contains(t, mock.killed, mockVideoContainerID)
	assert.Contains(t, mock.removed, mockVideoContainerID)
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_VideoCreateError(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	mock.failVideoCreate = true
	defer mock.server.Close()

	caps := session.Caps{Headless: true, Video: true, VideoName: "test.mp4"}
	_, err := newPlaywrightDockerStarter(t, mock, caps, testPlaywrightEnvironment()).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start video container")

	mock.mu.Lock()
	assert.Equal(t, 2, mock.createCalls)
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_VideoStartError(t *testing.T) {
	pwSrv := newPlaywrightHTTPServer()
	defer pwSrv.Close()

	mock := newPlaywrightDockerMock(t, pwSrv)
	mock.failStartVideo = true
	defer mock.server.Close()

	caps := session.Caps{Headless: true, Video: true, VideoName: "test.mp4"}
	_, err := newPlaywrightDockerStarter(t, mock, caps, testPlaywrightEnvironment()).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start video container")

	mock.mu.Lock()
	assert.Equal(t, 2, mock.createCalls)
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	assert.Contains(t, mock.removed, mockVideoContainerID)
	mock.mu.Unlock()
}

func TestPlaywrightDockerStartWithCancel_WaitServerErrorWithVideo(t *testing.T) {
	mock := newPlaywrightDockerMockWithPort(t, "127.0.0.1:1")
	defer mock.server.Close()

	env := testPlaywrightEnvironment()
	env.StartupTimeout = 200 * time.Millisecond

	caps := session.Caps{Headless: true, Video: true, VideoName: "test.mp4"}
	_, err := newPlaywrightDockerStarter(t, mock, caps, env).StartWithCancel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wait playwright server")

	mock.mu.Lock()
	assert.Contains(t, mock.killed, mockVideoContainerID)
	assert.Contains(t, mock.removed, mockVideoContainerID)
	assert.Contains(t, mock.removed, mockBrowserContainerID)
	mock.mu.Unlock()
}
