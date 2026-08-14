# ADR: Hub-attach warm Chrome WebDriver via `-warm-pool-url`

**Status:** accepted (2026-08-13)  
**Repo:** [qa-guru/selenoid](https://github.com/qa-guru/selenoid)  
**Companion:** [HUB-ATTACH.md](HUB-ATTACH.md) · [qa-guru/selenoid-warm-pool](https://github.com/qa-guru/selenoid-warm-pool)

## Pools

Three named pools. This ADR covers **warm** hub-attach only. **Cold** is the default, already-live path (hub `docker run` from `browsers.json`). **Hot** (reuse-session / preopen / bypass hub) is backlog window 03 — not this ADR.

| Pool | Role |
|------|------|
| **Cold** | Default. Existing Docker starter. Unchanged. |
| **Warm** | This ADR: Chrome WD attach to a pre-started loopback slot. |
| **Hot** | Same session + live page, client bypasses the hub. Not this ADR. |

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

- Local: `config.local.yaml` publishes `127.0.0.1:14441` (headed `:149`) and `14442` (`:149-min`, shm 2g) plus PW WS `14501/14502` — WD attach works when those ports listen.
- Box1 hub compose: `webdriver_url_loopback` + publish `127.0.0.1:14441→4444` (`qaguru/webdriver-chrome:149`, shm 2g) and `14442` (`:149-min`, shm 2g). Attach live on [selenoid.qa.guru](https://selenoid.qa.guru) for Chrome WD without video/VNC/HAR. Headed slot is first so the warm-pool Jenkins job prefers it.
- Box2 Jenkins: omit `loopback` (default false) — docker-DNS URLs unchanged.
- Tests: hub `warm/` + `service` unit; orchestrator reserve/loopback; pyramid does **not** hit a live pool (OUT).

## 2026-08-14 — warm 4/4 (window 02)

**Playwright hub-attach stays out.** Warm 4/4 means four containers up (`webdriver-chrome:149`, `:149-min`, `playwright-chromium:1.61.1`, `:1.61.1-min`). Hub `Find()` attach remains Chrome WD only (`warmEligible`). Playwright sessions still **cold**-start. PW slots are listed in the orchestrator (metrics + window 03 WS attach bypassing the hub). A later ADR change for PW hub-attach would be a hub cut (`FindPlaywright` + loopback WS), not compose.

**`-min` WD:** shm **2g** (256m died on New Session on box1). No extra client headless caps — the `-min` image is already headless.

**Moved to window 03 (hot):** reuse-session, `PREOPEN_URL`, `WarmRemote`, skip `open()`, JVM-keep / `closeAfterAll` off the wall, frontend refresh on GitHub deploy. Not this ADR.
