# Hub-attach: warm Chrome WebDriver

Hub can reuse a pre-started Chrome node from [selenoid-warm-pool](https://github.com/qa-guru/selenoid-warm-pool) instead of `docker run` on every session.

**ADR:** [ADR-hub-attach.md](ADR-hub-attach.md)

## Enable

Same flag as UI WARM metrics:

```bash
./selenoid -conf config/browsers.json -warm-pool-url http://127.0.0.1:9090
# or: SELENOID_WARM_POOL_URL=http://127.0.0.1:9090
```

Empty URL → no probe, no attach (cold Docker only).

## Flow

```
POST /wd/hub/session  (browserName=chrome, no video/VNC/HAR)
        │
        ├─ POST {warm-pool}/pool/reserve
        │    {protocol:webdriver, browser:chrome, owner:hub-<id>, loopback:true}
        │
        ├─ 200 + loopback webdriverUrl → proxy New Session to that ChromeDriver
        │    session end → POST /pool/release  (slot stays up)
        │
        └─ 409 / error / wait fail → cold Docker (unchanged)
```

Hub **always** asks for loopback URLs. The orchestrator reserves only slots whose WebDriver URL is reachable on the host (`127.0.0.1` / `localhost` / `::1`), either:

```yaml
webdriver_url: http://127.0.0.1:14441/          # already loopback
# or:
webdriver_url: http://warm-chrome-1:4444/       # docker-DNS for in-network clients
webdriver_url_loopback: http://127.0.0.1:14441/ # hub-on-host
```

If no such slot exists → **409** → cold. Box1 publishes `127.0.0.1:14441/14442` so attach is live; metrics stay on the same `-warm-pool-url`.

## Local stand

```bash
# slots (published WD ports) — once
docker compose -f docker-compose.local.yml up -d   # in selenoid-warm-pool/

# orchestrator — reuse stand, do not kill
python scripts/stands/ensure.py selenoid-warm-pool
curl -sf http://127.0.0.1:9090/health
curl -sf http://127.0.0.1:9090/pool/slots

# hub
./selenoid -conf config/browsers.json -warm-pool-url http://127.0.0.1:9090
```

Materials / URL gate: **GET** `/`, `/health`, `/pool/slots` only. Do not put `POST /pool/*` in Materials.

## Cold fallback (always)

| Condition | Result |
|-----------|--------|
| No `-warm-pool-url` | Cold |
| Not Chrome WD (firefox, PW, …) | Cold |
| `enableVideo` / `enableVNC` / `enableHAR` | Cold |
| Orchestrator 409 / down | Cold |
| Returned URL is not loopback | Release if reserved, then cold |
| ChromeDriver wait fail (~2s) | Release, then cold |

## Prod ([selenoid.qa.guru](https://selenoid.qa.guru))

Hub pin **v3.0.9** + warm-pool **v1.1.2** (attach since [v1.1.1](https://github.com/qa-guru/selenoid-warm-pool/releases/tag/v1.1.1)): loopback WD ports on `127.0.0.1:14441/14442` (`qaguru/webdriver-chrome:149`, shm 2g). Chrome WD sessions **without** video/VNC/HAR attach a warm slot; otherwise cold Docker. Ports are host-loopback only (not public).

Still out: Jenkins preopen / reuse-session · MCP · nginx `/pool/*` · Playwright slots · Box2 Jenkins jobs · Gridlane · UI changes · killing the local warm-pool stand.

## Verify

```bash
cd projects/selenoid-home/selenoid && go test ./warm/ ./service/ -count=1
cd projects/selenoid-home/selenoid-warm-pool && go test . -count=1
# live pyramid slice (stand :9090; hub-attach skips unless slots + hub -warm-pool-url)
python scripts/stands/ensure.py selenoid-warm-pool
cd projects/selenoid-home/selenoid-tests && ./scripts/run-go-pyramid.sh warm-pool
```
