# Release v2.0.0 — qa-guru/selenoid

**Дата:** 25 июня 2026  
**База:** форк [aerokube/selenoid](https://github.com/aerokube/selenoid) v1.11.x  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.0

Первый публичный релиз qa-guru с **нативной поддержкой Playwright** поверх классического Selenium WebDriver.

---

## Кратко

| | |
|---|---|
| **WebDriver** | Chrome 146–148 через `qaguru/webdriver-chrome` (ранее — legacy `twilio/selenoid` cold-образы) |
| **Playwright** | playwright-chromium, playwright-firefox, playwright-webkit 1.60–1.61 через `qaguru/playwright-chromium`, `playwright-firefox`, `playwright-webkit` |
| **Протокол PW** | WebSocket `ws://host:4444/playwright/{browser}/{version}` |
| **Видео / VNC** | Sidecar `qaguru/video-recorder`, VNC в образах qaguru/playwright-* |
| **Docker Hub** | `qaguru/selenoid:v2.0.0`, `qaguru/selenoid:latest-release` |
| **Бинарники** | `selenoid_linux_amd64`, `selenoid_darwin_arm64`, … |

---

## Новое в v2.0.0

### Native Playwright

- WebSocket-эндпоинт в стиле Moon: `/playwright/<browser>/<version>`
- Поддерживаемые браузеры в hub: `playwright-chromium`, `playwright-firefox`, `playwright-webkit`
- Поле `"protocol": "playwright"` в `browsers.json`
- Запуск browser-node через Docker с образами `qaguru/playwright-chromium`, `playwright-firefox`, …
- Query-параметры: `enableVideo`, `enableVNC`, `headless`, `name`, `videoName`, `screenResolution`, …
- Примеры клиентов: `@playwright/test`, `playwright` library — см. [playwright.md](playwright.md)

### browsers.json

Встроенный конфиг ([`config/browsers.json`](../config/browsers.json)) — два стека:

| Стек | Браузеры | Образы |
|------|----------|--------|
| WebDriver | `chrome` | `qaguru/webdriver-chrome:<major>` |
| Playwright | `playwright-chromium`, `playwright-firefox`, `playwright-webkit` | `qaguru/playwright-*:<version>` |

Таблица совместимости: [browser-versions.md](browser-versions.md).

### Документация и примеры

- [docs/playwright.md](playwright.md) — подключение клиентов, параметры сессии
- [docs/browser-versions.md](browser-versions.md) — матрица версий
- [examples/playwright/](../examples/playwright/) — smoke-тест на `@playwright/test`

### CI / релиз

- GitHub Release → сборка `gox`, загрузка ассетов в Release
- Docker-образ публикуется как **`qaguru/selenoid`** (нужны secrets `DOCKER_USERNAME` / `DOCKER_PASSWORD`)
- Go 1.23 в release workflow

---

## Установка

### Бинарник (Linux amd64)

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.0/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
```

### Docker

```bash
docker pull qaguru/selenoid:v2.0.0
# или latest-release
```

### Через cm (qa-guru fork)

```bash
cm selenoid start -v v2.0.0
```

---

## Миграция с aerokube/selenoid

1. Заменить образ hub на `qaguru/selenoid:v2.0.0` или бинарник из Release.
2. Обновить `browsers.json` — добавить секции Playwright и сменить WebDriver-образы на `qaguru/webdriver-chrome` (см. [`config/browsers.json`](../config/browsers.json)).
3. Подтянуть образы:
   ```bash
   docker pull qaguru/webdriver-chrome:148 qaguru/webdriver-chrome:148-min
   docker pull qaguru/playwright-chromium:1.61.1
   docker pull qaguru/video-recorder:latest
   ```
4. Для Playwright за nginx — проксировать WebSocket `/playwright/` через UI на hub.

**Обратная совместимость:** существующие Selenium-тесты на `/wd/hub` работают без изменений.

---

## Известные ограничения

- Версия npm-клиента `@playwright/test` должна совпадать с `playwrightVersion` в `browsers.json`.
- Образ `qaguru/selenoid` на Docker Hub появляется только при настроенных CI secrets.
- Документация на gh-pages (шаг `ci/docs.sh`) для форка опциональна и может быть пропущена.

---

## Связанные репозитории

| Репозиторий | Роль |
|-------------|------|
| [qa-guru/selenoid-ui](https://github.com/qa-guru/selenoid-ui) v2.0.0 | UI с Playwright в Capabilities |
| [qa-guru/browser-image](https://github.com/qa-guru/browser-image) | Docker-образ browser node |
