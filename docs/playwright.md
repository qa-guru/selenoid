# Native Playwright support

Selenoid fork adds first-class Playwright support via WebSocket endpoint (Moon-compatible URL scheme).

## Endpoint

```
ws://<selenoid-host>:4444/playwright/<browser>/<version>?<options>
```

Examples:

```
ws://localhost:4444/playwright/chromium/1.52.0
ws://localhost:4444/playwright/chromium/1.52.0?name=smoke&enableVideo=true
wss://selenoid.example.com/playwright/webkit/1.52.0
```

## Client setup

### @playwright/test

```bash
export PW_TEST_CONNECT_WS_ENDPOINT=ws://localhost:4444/playwright/chromium/1.52.0
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

const browser = await chromium.connect('ws://localhost:4444/playwright/chromium/1.52.0');
```

## browsers.json

Playwright browsers use `"protocol": "playwright"` and `mcr.microsoft.com/playwright` images:

```json
"chromium": {
  "default": "1.52.0",
  "versions": {
    "1.52.0": {
      "image": "mcr.microsoft.com/playwright:v1.52.0-noble",
      "port": "3000",
      "path": "/",
      "protocol": "playwright",
      "playwrightVersion": "1.52.0",
      "user": "pwuser",
      "workDir": "/home/pwuser",
      "shmSize": 2147483648
    }
  }
}
```

**Important:** Playwright client version must match `playwrightVersion` (major.minor).

## Query parameters

| Parameter | Description |
|-----------|-------------|
| `name` | Session name (label) |
| `enableVNC` | Enable VNC (future) |
| `enableVideo` | Enable video (future) |
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
docker pull mcr.microsoft.com/playwright:v1.52.0-noble
./selenoid -conf config/browsers.json -limit 5
```
