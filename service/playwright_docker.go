package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/info"
	"github.com/aerokube/selenoid/session"
	ctr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	playwrightProtocol = "playwright"
	defaultPWWorkDir   = "/home/pwuser"
	defaultPWUser      = "pwuser"
)

// PlaywrightDocker starts containers running Playwright browser server.
type PlaywrightDocker struct {
	ServiceBase
	Environment
	session.Caps
	LogConfig *ctr.LogConfig
	Client    *client.Client
}

// StartWithCancel implements Starter for native Playwright sessions.
func (d *PlaywrightDocker) StartWithCancel() (*StartedService, error) {
	portConfig, err := getPlaywrightPortConfig(d.Service, d.Caps, d.Environment)
	if err != nil {
		return nil, fmt.Errorf("configuring playwright ports: %v", err)
	}

	mem, err := getMemory(d.Service, d.Environment)
	if err != nil {
		return nil, fmt.Errorf("invalid memory limit: %v", err)
	}
	cpu, err := getCpu(d.Service, d.Environment)
	if err != nil {
		return nil, fmt.Errorf("invalid CPU limit: %v", err)
	}

	requestId := d.RequestId
	image := d.Service.Image.(string)
	ctx := context.Background()
	log.Printf("[%d] [CREATING_PLAYWRIGHT_CONTAINER] [%s]", requestId, image)

	hostConfig := ctr.HostConfig{
		Binds:        d.Service.Volumes,
		AutoRemove:   true,
		PortBindings: portConfig.PortBindings,
		LogConfig:    getLogConfig(*d.LogConfig, d.Caps),
		NetworkMode:  ctr.NetworkMode(d.Network),
		Tmpfs:        d.Service.Tmpfs,
		ShmSize:      getShmSize(d.Service),
		Privileged:   d.Privileged,
		Resources: ctr.Resources{
			Memory:   mem,
			NanoCPUs: cpu,
		},
		ExtraHosts: getExtraHosts(d.Service, d.Caps),
	}
	hostConfig.PublishAllPorts = d.Service.PublishAllPorts
	if len(d.Caps.DNSServers) > 0 {
		hostConfig.DNS = d.Caps.DNSServers
	}

	pwVersion := d.Service.PlaywrightVersion
	if pwVersion == "" {
		pwVersion = "latest"
	}
	port := d.Service.Port
	runServerCmd := buildPlaywrightStartupScript(pwVersion, port, d.Caps.Name, d.Caps)

	cfg := &ctr.Config{
		Image:        image,
		Cmd:          []string{"/bin/sh", "-c", runServerCmd},
		Env:          getEnv(d.ServiceBase, d.Caps),
		ExposedPorts: portConfig.ExposedPorts,
		Labels:       getLabels(d.Service, d.Caps),
	}
	if user := d.Service.User; user != "" {
		cfg.User = user
	} else {
		cfg.User = defaultPWUser
	}
	if workDir := d.Service.WorkDir; workDir != "" {
		cfg.WorkingDir = workDir
	} else {
		cfg.WorkingDir = defaultPWWorkDir
	}
	if hn := getContainerHostname(d.Caps); hn != "" {
		cfg.Hostname = hn
	}

	container, err := d.Client.ContainerCreate(ctx, cfg, &hostConfig, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create playwright container: %v", err)
	}

	browserContainerStartTime := time.Now()
	browserContainerId := container.ID
	videoContainerId := ""
	log.Printf("[%d] [STARTING_PLAYWRIGHT_CONTAINER] [%s] [%s]", requestId, image, browserContainerId)
	if err = d.Client.ContainerStart(ctx, browserContainerId, ctr.StartOptions{}); err != nil {
		removeContainer(ctx, d.Client, requestId, browserContainerId)
		return nil, fmt.Errorf("start playwright container: %v", err)
	}
	log.Printf("[%d] [PLAYWRIGHT_CONTAINER_STARTED] [%s] [%s] [%.2fs]", requestId, image, browserContainerId, info.SecondsSince(browserContainerStartTime))

	stat, err := waitContainerInspect(ctx, d.Client, browserContainerId, portConfig.SeleniumPort, d.StartupTimeout)
	if err != nil {
		removeContainer(ctx, d.Client, requestId, browserContainerId)
		return nil, err
	}

	servicePort := d.Service.Port
	pc := map[string]nat.Port{
		servicePort: portConfig.SeleniumPort,
	}
	if portConfig.VNCPort != "" {
		pc[ports.VNC] = portConfig.VNCPort
	}
	hostPort := getHostPort(d.Environment, servicePort, d.Caps, stat, pc)
	if hostPort.Playwright == "" {
		removeContainer(ctx, d.Client, requestId, browserContainerId)
		return nil, fmt.Errorf("no bindings available for playwright port %s", servicePort)
	}

	if d.Caps.Video {
		videoContainerId, err = startVideoContainer(ctx, d.Client, requestId, stat, d.Environment, d.ServiceBase, d.Caps)
		if err != nil {
			removeContainer(ctx, d.Client, requestId, browserContainerId)
			return nil, fmt.Errorf("start video container: %v", err)
		}
	}

	servicePath := d.Service.Path
	if servicePath == "" {
		servicePath = "/"
	}
	wsURL := &url.URL{Scheme: "ws", Host: hostPort.Playwright, Path: servicePath}

	serviceStartTime := time.Now()
	if err = waitPlaywrightServer(hostPort.Playwright, d.StartupTimeout); err != nil {
		if videoContainerId != "" {
			stopVideoContainer(ctx, d.Client, requestId, videoContainerId, d.Environment)
		}
		removeContainer(ctx, d.Client, requestId, browserContainerId)
		return nil, fmt.Errorf("wait playwright server: %v", err)
	}
	log.Printf("[%d] [PLAYWRIGHT_SERVICE_STARTED] [%s] [%s] [%.2fs]", requestId, image, browserContainerId, info.SecondsSince(serviceStartTime))
	log.Printf("[%d] [PLAYWRIGHT_PROXY_TO] [%s] [%s]", requestId, browserContainerId, wsURL.String())

	return &StartedService{
		Url: wsURL,
		Container: &session.Container{
			ID:        browserContainerId,
			IPAddress: getContainerIP(d.Environment.Network, stat),
		},
		HostPort: hostPort,
		Cancel: func() {
			if videoContainerId != "" {
				stopVideoContainer(ctx, d.Client, requestId, videoContainerId, d.Environment)
			}
			removeContainer(ctx, d.Client, requestId, browserContainerId)
		},
	}, nil
}

func buildPlaywrightStartupScript(pwVersion, port, browserName string, caps session.Caps) string {
	runServer := fmt.Sprintf("exec node /home/pwuser/node_modules/playwright/cli.js run-server --port %s --host 0.0.0.0", port)
	if !caps.VNC && !caps.Video {
		return runServer
	}

	resolution := caps.ScreenResolution
	if resolution == "" {
		resolution = "1920x1080x24"
	}

	vncStart := ""
	if caps.VNC {
		vncStart = fmt.Sprintf("x11vnc -display :99 -forever -shared -rfbport %s -passwd selenoid >/dev/null 2>&1 &\n", ports.VNC)
	}

	headedLauncher := ""
	if caps.VNC && !caps.Headless && caps.TestName == "Manual session" {
		pwBrowser := playwrightBrowserName(browserName)
		headedLauncher = fmt.Sprintf(`(sleep 2; PW_BROWSER=%s PW_PORT=%s node /home/pwuser/launch-headed-browser.js >>/tmp/headed-launch.log 2>&1 &)
`, pwBrowser, port)
	}

	return fmt.Sprintf(`set -e
Xvfb :99 -screen 0 %s -ac +extension RANDR -noreset -listen tcp >/dev/null 2>&1 &
export DISPLAY=:99
for i in $(seq 1 50); do xdpyinfo -display :99 >/dev/null 2>&1 && break; sleep 0.1; done
%s%s%s`, resolution, vncStart, headedLauncher, runServer)
}

func playwrightBrowserName(browserName string) string {
	switch strings.ToLower(browserName) {
	case "webkit":
		return "webkit"
	case "firefox-playwright":
		return "firefox"
	default:
		return "chromium"
	}
}

func getPlaywrightPortConfig(service *config.Browser, caps session.Caps, env Environment) (*portConfig, error) {
	serverPort, err := nat.NewPort("tcp", service.Port)
	if err != nil {
		return nil, fmt.Errorf("new playwright port: %v", err)
	}
	exposedPorts := map[nat.Port]struct{}{serverPort: {}}
	portBindings := nat.PortMap{}
	if env.IP != "" || !env.InDocker {
		portBindings[serverPort] = []nat.PortBinding{{HostIP: "0.0.0.0"}}
	}

	var vnc nat.Port
	if caps.VNC {
		vnc, err = nat.NewPort("tcp", ports.VNC)
		if err != nil {
			return nil, fmt.Errorf("new vnc port: %v", err)
		}
		exposedPorts[vnc] = struct{}{}
		if env.IP != "" || !env.InDocker {
			portBindings[vnc] = []nat.PortBinding{{HostIP: "0.0.0.0"}}
		}
	}

	return &portConfig{
		SeleniumPort: serverPort,
		VNCPort:      vnc,
		PortBindings: portBindings,
		ExposedPorts: exposedPorts,
	}, nil
}

func waitPlaywrightServer(hostPort string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	url := "http://" + hostPort + "/"
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("%s does not respond in %v", hostPort, timeout)
}

func waitTCP(hostPort string, t time.Duration) error {
	deadline := time.Now().Add(t)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", hostPort, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("%s does not respond in %v", hostPort, t)
}

func isPlaywrightBrowser(service *config.Browser) bool {
	return strings.EqualFold(service.Protocol, playwrightProtocol)
}
