# Container-reuse: warm Chrome WebDriver

Hub can reuse a pre-started Chrome **container** from [selenoid-pool](https://github.com/qa-guru/selenoid-pool) instead of `docker run` on every session.

**ADR:** [ADR-container-reuse.md](ADR-container-reuse.md)  
Not an Allure **attachment**.

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

On a clean Docker host, [qa-guru/cm](https://github.com/qa-guru/cm) starts the sidecar and points the hub at it:

```bash
./cm selenoid start -c "$HOME/selenoid" --warm-pool
# alias: --pool
curl -sf http://127.0.0.1:9090/health
curl -sf http://127.0.0.1:9090/pool/slots
# hub /status → warmTotal>0
```

Hub already running? Pass `-warm-pool-url http://127.0.0.1:9090` (or `SELENOID_WARM_POOL_URL`).

Materials / URL gate: **GET** `/`, `/health`, `/pool/slots` only. Do not put `POST /pool/*` in Materials.

### Monorepo (zero-design-system)

Hub developers in this workspace — slots + orchestrator without cm:

```bash
# slots (published WD ports) — once
docker compose -f docker-compose.local.yml up -d   # in selenoid-pool/

python scripts/stands/ensure.py selenoid-pool
curl -sf http://127.0.0.1:9090/health

./selenoid -conf config/browsers.json -warm-pool-url http://127.0.0.1:9090
```

## Pool selector (client contract)

Sessions pick a pool explicitly via `selenoid:options.pool` or `labels.pool`
(`selenoid:options` wins when both are set):

| Value | Behavior |
|-------|----------|
| unset / `auto` | Prefer warm; cold fallback possible — the response says which |
| `cold` | Always cold-start (documented replacement for the `enableHar=true` workaround) |
| `warm` | Strict: warm slot or `session not created` — never silent cold |
| `hot` | `invalid argument` — hot slots live behind `POST /pool/lease`, not hub `/session` |
| other | `invalid argument` naming the bad value |

The New Session response carries the actual routing in `value.capabilities`:

```json
"servedPool": "warm",            // warm | cold — who really serves it
"requestedPool": "auto",         // normalized selector (auto when unset)
"servedSlot": "pool-chrome-1",   // warm only
"forcedPoolReason": "sidecar-caps"  // only when rerouted away from warm
```

`forcedPoolReason` tokens: `sidecar-caps` (enableVideo/enableVNC/enableHAR),
`non-chrome`, `version-pin`, `min-image`, `warm-unavailable`,
`warm-attach-failed`, `no-warm-pool`. With `pool=warm` the same conditions
refuse the session instead of falling back — the error message names
`forcedPoolReason=<token>`.

Hub `/status` aggregates it: `poolRouting.warmServed` / `coldServed` /
`warmFallbacks` counters plus `poolRouting.recent[]` (last 20 routings), and
each slot row shows `reservedBy`, `reservedAt`, `leaseExpiresAt`,
`sessionCount`, `ageSec`, `recycling`, `recyclePending`.

## Lease and release (orchestrator ≥ v1.2)

Reserve/lease write `reservedBy` + `reservedAt` + `leaseExpiresAt` on the slot.
`release` and `renew` must echo the reserving `owner` — a mismatch is a 409;
`"force":true` is the operator bypass; reservations without an owner are
`anonymous` (anyone may release). The hub reserves as `hub-<requestId>` with
`ttlSec = -max-timeout + 30m` and releases with the same owner on session end.

A reaper sweeps expired leases every 15s (`lease_ttl` default 1h,
`max_lease_ttl` clamp 6h) — a dead agent cannot hold a slot forever. Warm
containers rotate on `recycle_after_sessions` / `recycle_after_age` via
`docker restart` (needs the socket mount); a rotating slot reads
`recycling:true` and is not handed out, a failed rotation sets
`recyclePending:true` and keeps serving.

## Cold fallback (always)

| Condition | Result |
|-----------|--------|
| No `-warm-pool-url` | Cold (`forcedPoolReason=no-warm-pool`) |
| Not Chrome WD (firefox, PW, …) | Cold (`non-chrome`) |
| `enableVideo` / `enableVNC` / `enableHAR` | Cold (`sidecar-caps`) — with `pool=warm` a refusal instead |
| Compat pin / `-min` version | Cold (`version-pin` / `min-image`) |
| Orchestrator 409 / down | Cold (`warm-unavailable`) |
| Returned URL is not loopback | Release if reserved, then cold (`warm-unavailable`) |
| ChromeDriver wait fail (~2s) | Release, then cold (`warm-attach-failed`) |

`pool=cold` skips the pool entirely. `pool=warm` is strict — every row above
becomes a `session not created` error instead of a cold fallback.

## Prod ([selenoid.qa.guru](https://selenoid.qa.guru))

Hub pin **v3.0.9** + pool **v1.1.2** (container-reuse since [v1.1.1](https://github.com/qa-guru/selenoid-pool/releases/tag/v1.1.1)). Compose SSOT is warm **4/4**: headed `qaguru/webdriver-chrome:149` on `127.0.0.1:14441` (shm 2g, first reserve), `:149-min` on `14442` (shm 2g), Playwright `1.61.1` / `1.61.1-min` on `14501/14502`. Live box1 until that compose is deployed may still be 2× `:149`. Chrome WD sessions **without** video/VNC/HAR reuse a warm WD container; Playwright and everything else stay **cold** Docker. Ports are host-loopback only (not public).

Still out: Jenkins preopen / session-reuse · MCP · nginx `/pool/*` · Playwright **container-reuse** · Box2 Jenkins jobs · Gridlane · UI changes · killing the local warm-pool stand.

## Verify

Public clone of this repo:

```bash
go test ./warm/ ./service/ -count=1
```

Live slots after `cm selenoid start --warm-pool`: Chrome WD to `http://127.0.0.1:4444/wd/hub` (no video/VNC/HAR) should reuse; otherwise cold fallback.

### Monorepo (zero-design-system)

```bash
cd projects/selenoid-home/selenoid && go test ./warm/ ./service/ -count=1
cd projects/selenoid-home/selenoid-pool && go test . -count=1
# live pyramid slice (stand :9090; container-reuse skips unless slots + hub -warm-pool-url)
python scripts/stands/ensure.py selenoid-pool
cd projects/selenoid-home/selenoid-tests && ./scripts/run-go-pyramid.sh warm-pool
```
