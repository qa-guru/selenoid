# Container-reuse: warm Chrome WebDriver

Hub can reuse a pre-started Chrome **container** from [selenoid-warm-pool](https://github.com/qa-guru/selenoid-warm-pool) instead of `docker run` on every session.

**ADR:** [ADR-hub-attach.md](ADR-hub-attach.md)  
**Former name:** hub-attach (filename kept). Not an Allure **attachment**.

## Glossary

| Term | Pool | Means |
|------|------|--------|
| **container-reuse** | warm (this doc) | New Session on a container that is already up |
| **session-reuse** | hot | Client joins an existing WD UUID / PW WS; bypasses the hub |
| **cold** | cold | Hub `docker run` + New Session |
| Allure **attachment** | — | Report file (screenshot / page source / console). Jobs `*-full-attachments` can run on warm; that is not session-reuse. |

## Pools

| Pool | What | Client |
|------|------|--------|
| **Cold** (default, live) | Classic Selenoid: hub `docker run` of the image from `browsers.json` for each session | Hub `POST /session` |
| **Warm** (this doc) | Container already up; **container-reuse** = New Session on that node | Hub → `POST /pool/reserve` |
| **Hot** | Same WD/PW session + live page; **session-reuse** by UUID / WS **bypassing the hub** | Backlog — window 03 |

Cold is the existing path: no `-warm-pool-url`, 409, not Chrome WD, or video/VNC/HAR. Do not change the Docker starter.

## Enable

Same flag as UI WARM metrics:

```bash
./selenoid -conf config/browsers.json -warm-pool-url http://127.0.0.1:9090
# or: SELENOID_WARM_POOL_URL=http://127.0.0.1:9090
```

Empty URL → no probe, no container-reuse (cold Docker only).

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

If no such slot exists → **409** → cold. Box1 publishes `127.0.0.1:14441/14442` so container-reuse is live; metrics stay on the same `-warm-pool-url`.

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

Hub pin **v3.0.9** + warm-pool **v1.1.2** (container-reuse since [v1.1.1](https://github.com/qa-guru/selenoid-warm-pool/releases/tag/v1.1.1)). Compose SSOT is warm **4/4**: headed `qaguru/webdriver-chrome:149` on `127.0.0.1:14441` (shm 2g, first reserve), `:149-min` on `14442` (shm 2g), Playwright `1.61.1` / `1.61.1-min` on `14501/14502`. Live box1 until that compose is deployed may still be 2× `:149`. Chrome WD sessions **without** video/VNC/HAR reuse a warm WD container; Playwright and everything else stay **cold** Docker. Ports are host-loopback only (not public).

Still out: Jenkins preopen / session-reuse · MCP · nginx `/pool/*` · Playwright **container-reuse** · Box2 Jenkins jobs · Gridlane · UI changes · killing the local warm-pool stand.

## Verify

```bash
cd projects/selenoid-home/selenoid && go test ./warm/ ./service/ -count=1
cd projects/selenoid-home/selenoid-warm-pool && go test . -count=1
# live pyramid slice (stand :9090; container-reuse skips unless slots + hub -warm-pool-url)
python scripts/stands/ensure.py selenoid-warm-pool
cd projects/selenoid-home/selenoid-tests && ./scripts/run-go-pyramid.sh warm-pool
```
