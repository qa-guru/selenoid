# Native Playwright

Форк Selenoid добавляет нативную поддержку Playwright через WebSocket endpoint (схема URL совместима с Moon).

## Endpoint

```
ws://<selenoid-host>:4444/playwright/<browser>/<version>?<options>
```

Примеры:

```
ws://localhost:4444/playwright/playwright-chromium/1.61.1
ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=smoke&enableVideo=true
wss://selenoid.example.com/playwright/playwright-webkit/1.61.1
```

Production hub (`selenoid.autotests.cloud`): [qa-guru/cm deploy/README.md](https://github.com/qa-guru/cm/blob/main/deploy/README.md).

## Настройка клиента

### @playwright/test

```bash
export PW_TEST_CONNECT_WS_ENDPOINT=ws://localhost:4444/playwright/playwright-chromium/1.61.1
npx playwright test
```

Или в `playwright.config.ts`:

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

const browser = await chromium.connect('ws://localhost:4444/playwright/playwright-chromium/1.61.1');
```

## browsers.json

Playwright-браузеры используют `"protocol": "playwright"` и отдельные образы на каждый браузер:

```json
"playwright-chromium": {
  "default": "1.61.1",
  "versions": {
    "1.61.1": {
      "image": "qaguru/playwright-chromium:1.61.1",
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

Образы: `qaguru/playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge`.

**Важно:** версия Playwright-клиента должна совпадать с `playwrightVersion` (major.minor.patch).

Полная матрица совместимости (Playwright ↔ движки Chromium/Firefox/WebKit, WebDriver-образы, Docker-теги) — в [browser-versions.md](browser-versions.md).

## Query-параметры

| Параметр | Описание |
|----------|----------|
| `name` | Имя сессии (метка) |
| `enableVNC` | Живой экран браузера через VNC (нужен образ `qaguru/playwright-*`) |
| `headless` | Запуск с UI (`false`) для VNC; по умолчанию `true`. Ручные сессии через UI передают `headless=false` автоматически |
| `enableVideo` | Запись сессии в H.264 (нужен образ `selenoid/video-recorder`) |
| `videoName` | Имя видеофайла (например `smoke.mp4`); по умолчанию `<session-id>.mp4` |
| `enableLog` | Сохранять логи контейнера |
| `screenResolution` | Например `1920x1080x24` |
| `sessionTimeout` | Например `5m` |
| `timeZone` | Часовой пояс контейнера |
| `env.KEY` | Дополнительные переменные окружения контейнера |

## Сборка

```bash
go build -o selenoid .
```

Или:

```bash
./scripts/build-selenoid.sh
```

## Локальный запуск

```bash
docker pull qaguru/playwright-chromium:1.61.1
# или сборка из https://github.com/qa-guru/playwright-image
docker pull selenoid/video-recorder:latest-release
./selenoid -conf config/browsers.json -limit 5
```

## Образ Playwright-браузера

Исходники и сборка образов — отдельный репозиторий [qa-guru/playwright-image](https://github.com/qa-guru/playwright-image) (`qaguru/playwright-*` на Docker Hub).

```bash
git clone https://github.com/qa-guru/playwright-image.git
cd playwright-image
./scripts/build.sh chromium 1.61.1
```

## Публикация в Docker Hub

См. [playwright-image](https://github.com/qa-guru/playwright-image): `docker login` и `./scripts/push.sh all 1.61.1`.

## Toolchain

Hub требует **Docker Engine 26.1.x** (API **1.45**) и **Go 1.23.x**. Локальный старт:

```bash
./scripts/start-selenoid.sh
```

См. также [RELEASE_v2.0.9.md](RELEASE_v2.0.9.md) и [docker-settings.adoc](docker-settings.adoc).
