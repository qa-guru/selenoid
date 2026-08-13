# ADR: Hub-attach warm Chrome WebDriver via `-warm-pool-url`

**Status:** accepted (2026-08-13)  
**Repo:** [qa-guru/selenoid](https://github.com/qa-guru/selenoid)  
**Companion:** [HUB-ATTACH.md](HUB-ATTACH.md) · [qa-guru/selenoid-warm-pool](https://github.com/qa-guru/selenoid-warm-pool)

## Context

`-warm-pool-url` / `SELENOID_WARM_POOL_URL` already probes the warm-pool orchestrator and exposes `warmReady` / `warmTotal` on hub `/status` (UI WARM). Session create still always started a **cold** Docker container.

Jenkins on box2 already talks to the orchestrator (`POST /pool/reserve`) and drives slots directly. Hub-side attach was the documented follow-up.

Hub is a **native host binary** (not hub-in-docker). Slot `webdriver_url` values like `http://warm-chrome-1:4444/` are docker-DNS — unreachable from the host.

## Decision

1. **Chrome WebDriver only.** `Find()` may attach a reserved warm slot instead of starting Docker. Playwright, Firefox, Edge, Android — unchanged (cold).
2. **Loopback or skip.** Hub always sends `"loopback": true` on `POST /pool/reserve`. The orchestrator only reserves a slot that has a loopback WebDriver URL (`webdriver_url_loopback` or `webdriver_url` whose host is `127.0.0.1` / `localhost` / `::1`). Otherwise **409** — hub uses cold Docker. Prod box1 compose without published WD ports stays metrics-only; no 2s wait per session.
3. **Cold fallback.** 409, orchestrator down, empty URL, wait failure, or Docker-only caps (`enableVideo` / `enableVNC` / `enableHAR`) → existing Docker starter. A reserved slot is released before falling back.
4. **Release, do not kill.** Session cancel → `POST /pool/release` (orchestrator best-effort `/warm/reset`). The slot container stays up.
5. **Same `-warm-pool-url`.** No second flag. Empty URL keeps today’s behaviour (no probe, no attach).

## Rejected

| Alternative | Why not |
|-------------|---------|
| Attach using docker-DNS URLs | Host hub cannot resolve `warm-chrome-N` |
| Opt-in capability `selenoid:options.warmPool` | Extra client contract; CI already speaks plain Chrome WD |
| Attach Playwright / video / VNC / HAR in this cut | Needs published 7070/5900 + sidecar; OUT of this slice |
| Separate `-warm-pool-attach` flag | User contract is one URL; loopback filter is the prod safety net |

## Consequences

- Local: `config.local.yaml` already publishes `127.0.0.1:14441/14442` — attach works when those ports listen.
- Box1 hub compose: `webdriver_url_loopback` + publish `127.0.0.1:14441/14442→4444` (`qaguru/webdriver-chrome:149-min`) — attach live on [selenoid.qa.guru](https://selenoid.qa.guru) for Chrome WD without video/VNC/HAR.
- Box2 Jenkins: omit `loopback` (default false) — docker-DNS URLs unchanged.
- Tests: hub `warm/` + `service` unit; orchestrator reserve/loopback; pyramid does **not** hit a live pool (OUT).
