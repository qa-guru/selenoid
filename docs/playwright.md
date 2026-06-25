# Native Playwright support

Selenoid fork adds first-class Playwright support via WebSocket endpoint (Moon-compatible URL scheme).

## Endpoint

```
ws://<selenoid-host>:4444/playwright/<browser>/<version>?<options>
```

Examples:

```
ws://localhost:4444/playwright/chromium/1.61.1
ws://localhost:4444/playwright/chromium/1.61.1?name=smoke&enableVideo=true
wss://selenoid.example.com/playwright/webkit/1.61.1
```

## Client setup

### @playwright/test

```bash
export PW_TEST_CONNECT_WS_ENDPOINT=ws://localhost:4444/playwright/chromium/1.61.1
npx playwright test
```

Or in `playwright.config.ts`:

```typescript
use: {
  connectOptions: {
    wsEndpoint: process.env.PLAYWRIGHT_WS_ENDPOINT,
  },
},
```

### playwright library

```typescript
import { chromium } from 'playwright';

const browser = await chromium.connect('ws://localhost:4444/playwright/chromium/1.61.1');
```

## browsers.json

Playwright browsers use `"protocol": "playwright"` and `mcr.microsoft.com/playwright` images:

```json
"chromium": {
  "default": "1.61.1",
  "versions": {
    "1.61.1": {
      "image": "qaguru/playwright:v1.61.1-noble",
      "port": "3000",
      "path": "/",
      "protocol": "playwright",
      "playwrightVersion": "1.61.1",
      "user": "pwuser",
      "workDir": "/home/pwuser",
      "shmSize": 2147483648
    }
  }
}
```

**Important:** Playwright client version must match `playwrightVersion` (major.minor.patch).

See [browser-versions.md](browser-versions.md) for the full compatibility matrix (Playwright ↔ Chromium/Firefox/WebKit engines, WebDriver images, Docker tags).

## Query parameters

| Parameter | Description |
|-----------|-------------|
| `name` | Session name (label) |
| `enableVNC` | Enable live browser screen via VNC (requires `qaguru/playwright` image) |
| `headless` | Run browser headed (`false`) for VNC; default `true`. Manual UI sessions pass `headless=false` automatically |
| `enableVideo` | Enable H.264 session recording (requires `selenoid/video-recorder` image) |
| `videoName` | Custom video file name (e.g. `smoke.mp4`); default is `<session-id>.mp4` |
| `enableLog` | Save container logs |
| `screenResolution` | e.g. `1920x1080x24` |
| `sessionTimeout` | e.g. `5m` |
| `timeZone` | Container timezone |
| `env.KEY` | Extra container env vars |

## Build

```bash
go build -o selenoid .
```

Or:

```bash
./scripts/build-selenoid.sh
```

## Run locally

```bash
docker pull qaguru/playwright:v1.61.1-noble
# или сборка из https://github.com/qa-guru/playwright-image
docker pull selenoid/video-recorder:latest-release
./selenoid -conf config/browsers.json -limit 5
```

## Playwright browser image

Исходники и сборка образа — отдельный репозиторий [qa-guru/playwright-image](https://github.com/qa-guru/playwright-image) (`qaguru/playwright` на Docker Hub).

```bash
git clone https://github.com/qa-guru/playwright-image.git
cd playwright-image
./scripts/build.sh v1.61.1-noble
```

## Publish to Docker Hub

См. [playwright-image](https://github.com/qa-guru/playwright-image): `docker login` и `./scripts/push.sh v1.61.1-noble`.

## Toolchain

Hub requires **Docker Engine 26.1.x** (API **1.45**) and **Go 1.23.x**. Start locally:

```bash
./scripts/start-selenoid.sh
```

See [RELEASE_v2.0.1.md](RELEASE_v2.0.1.md) and [docker-settings.adoc](docker-settings.adoc).

