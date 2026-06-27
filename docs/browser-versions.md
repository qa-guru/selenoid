# Таблица совместимости версий браузеров

Актуально для [`config/browsers.json`](../config/browsers.json). Последняя проверка: июнь 2026.

В Selenoid используются **два независимых стека**. Версии в `browsers.json` для них означают разное:

| Стек | Браузеры в `browsers.json` | Протокол | Что означает поле `version` |
|------|----------------------------|----------|------------------------------|
| **WebDriver** | `chrome`, `firefox`, `msedge` | Selenium (`/wd/hub`) | Мажорная версия браузера (`148.0` → Chrome 148.x) |
| **Playwright** | `playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` | WebSocket (`/playwright/...`) | Версия npm-пакета `@playwright/test` (`1.61.1`) |

Playwright-образы — per-browser на Docker Hub: `qaguru/playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` ([исходники](https://github.com/qa-guru/playwright-image)).

---

## Сводная таблица (`browsers.json`)

| Имя в hub | Default | Версии в конфиге | Docker-образ | Протокол |
|-----------|---------|------------------|--------------|----------|
| `chrome` | `148.0` | 148.0, 147.0, 146.0, 128.0 | `twilio/selenoid:chrome_stable_<N>`, `selenoid/vnc_chrome:128.0` | WebDriver |
| `firefox` | `150.0` | 150.0, 149.0, 148.0 | `twilio/selenoid:firefox_stable_<N>` | WebDriver |
| `msedge` | `145.0` | 145.0, 144.0, 143.0 | `twilio/selenoid:edge_stable_<N>` | WebDriver |
| `playwright-chromium` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0, 1.46.0 | `qaguru/playwright-chromium:<версия>` | Playwright |
| `playwright-firefox` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-firefox:<версия>` | Playwright |
| `playwright-webkit` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-webkit:<версия>` | Playwright |
| `playwright-chrome` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-chrome:<версия>` | Playwright |
| `playwright-msedge` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-msedge:<версия>` | Playwright |

---

## WebDriver: Chrome, Firefox и Edge (Twilio)

Источник: [hub.docker.com/r/twilio/selenoid](https://hub.docker.com/r/twilio/selenoid). Образы совместимы с Selenoid, обновляются регулярно (последние stable — май 2026).

Тег Docker (`chrome_stable_148`) ↔ capability `version: 148.0`.

### Chrome (`twilio/selenoid`)

| Версия в hub | Docker-тег | Chrome в контейнере | `path` | Endpoint |
|--------------|------------|---------------------|--------|----------|
| **148.0** *(default)* | `chrome_stable_148` | Chrome 148.x | `/` | `POST /wd/hub/session`, `browserName: chrome` |
| 147.0 | `chrome_stable_147` | Chrome 147.x | `/` | то же, `version: 147.0` |
| 146.0 | `chrome_stable_146` | Chrome 146.x | `/` | то же, `version: 146.0` |

### Firefox (`twilio/selenoid`)

| Версия в hub | Docker-тег | Firefox в контейнере | `path` | Endpoint |
|--------------|------------|----------------------|--------|----------|
| **150.0** *(default)* | `firefox_stable_150` | Firefox 150.x | `/wd/hub` | `POST /wd/hub/session`, `browserName: firefox` |
| 149.0 | `firefox_stable_149` | Firefox 149.x | `/wd/hub` | то же, `version: 149.0` |
| 148.0 | `firefox_stable_148` | Firefox 148.x | `/wd/hub` | то же, `version: 148.0` |

### Microsoft Edge (`twilio/selenoid`)

| Версия в hub | Docker-тег | Edge в контейнере | `path` | Endpoint |
|--------------|------------|-------------------|--------|----------|
| **145.0** *(default)* | `edge_stable_145` | Edge 145.x | `/` | `POST /wd/hub/session`, `browserName: msedge` (алиас `MicrosoftEdge`) |
| 144.0 | `edge_stable_144` | Edge 144.x | `/` | то же, `version: 144.0` |
| 143.0 | `edge_stable_143` | Edge 143.x | `/` | то же, `version: 143.0` |

> Edge в twilio/selenoid доступен только на **linux/amd64**. Теги 146+ пока отсутствуют (Chrome уже на 148).

**Клиент:** любой Selenium WebDriver. Версия Selenium-клиента не привязана к версии браузера.

> Устаревшие образы `selenoid/chrome` и `selenoid/firefox` (Aerokube, archived) больше не используются — максимум там Chrome 128 / Firefox 125.

---

## Playwright: движки внутри каждой версии

Данные из официального [`browsers.json`](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json) Microsoft Playwright.

### Chromium Playwright (`playwright-chromium`)

| Playwright в hub | `playwrightVersion` | Chromium в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|------------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **149.0.7827.55** (rev 1228) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-chromium/1.61.1` |
| 1.61.0 | `1.61.0` | **149.0.7827.55** (rev 1228) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/playwright-chromium/1.61.0` |
| 1.60.0 | `1.60.0` | **148.0.7778.96** (rev 1223) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/playwright-chromium/1.60.0` |

### Firefox Playwright (`playwright-firefox`)

| Playwright в hub | `playwrightVersion` | Firefox в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|----------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **151.0** (rev 1532) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-firefox/1.61.1` |
| 1.61.0 | `1.61.0` | **151.0** (rev 1532) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/playwright-firefox/1.61.0` |
| 1.60.0 | `1.60.0` | **150.0.2** (rev 1522) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/playwright-firefox/1.60.0` |

### WebKit Playwright (`playwright-webkit`)

| Playwright в hub | `playwrightVersion` | WebKit в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|---------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **26.5** (rev 2311) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-webkit/1.61.1` |
| 1.61.0 | `1.61.0` | **26.5** (rev 2311) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/playwright-webkit/1.61.0` |
| 1.60.0 | `1.60.0` | **26.4** (rev 2287) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/playwright-webkit/1.60.0` |

---

## Правила совместимости клиента

| Что проверять | Правило | Последствие при нарушении |
|---------------|---------|---------------------------|
| Playwright npm ↔ hub | `@playwright/test` **должен совпадать** с `playwrightVersion` и версией в URL | Ошибки протокола, несовместимость CDP/BiDi |
| Playwright ↔ WebDriver | **Нельзя** подставить `chrome:148.0` вместо `playwright-chromium:1.61.1` | Сессия не создастся |
| WebDriver chrome ↔ firefox | Независимы | — |
| Образ Docker | Образ из `browsers.json` должен быть **скачан** до старта сессии | `No such image` при создании сессии |

---

## Сравнение движков: Playwright vs WebDriver (Twilio)

| Playwright | Chromium/Firefox (Playwright) | WebDriver Twilio | Комментарий |
|------------|--------------------------------|------------------|-------------|
| 1.61.x | Chromium **149** | Chrome **148.0** | Почти на одном уровне |
| 1.61.x | Firefox **151** | Firefox **150.0** | Почти на одном уровне |
| 1.60.x | Chromium **148** | Chrome **148.0** | Совпадает по мажору |

---

## Матрица образов Docker

| Компонент | Команда pull / build |
|-----------|----------------------|
| Chrome 148/147/146 | `docker pull twilio/selenoid:chrome_stable_148` (и 147, 146) |
| Firefox 150/149/148 | `docker pull twilio/selenoid:firefox_stable_150` (и 149, 148) |
| Edge 145/144/143 | `docker pull twilio/selenoid:edge_stable_145` (и 144, 143) |
| Playwright 1.61.1 | [playwright-image](https://github.com/qa-guru/playwright-image) `./scripts/build.sh all 1.61.1` или `docker pull qaguru/playwright-chromium:1.61.1` |
| Playwright 1.61.0 | `./scripts/build.sh all 1.61.0` или `docker pull qaguru/playwright-chromium:1.61.0` |
| Playwright 1.60.0 | `./scripts/build.sh all 1.60.0` или `docker pull qaguru/playwright-chromium:1.60.0` |
| Видеозапись | `docker pull selenoid/video-recorder:latest-release` |

Базовый слой Playwright: `mcr.microsoft.com/playwright:v<версия>-noble` (Ubuntu Noble). Образы [`qaguru/playwright-*`](https://github.com/qa-guru/playwright-image) публикуются с тегом `<версия>` (например `1.61.1`).

Полный список тегов Twilio: [twilio/selenoid tags](https://hub.docker.com/r/twilio/selenoid/tags).

---

## Быстрый выбор стека

| Задача | Браузер в hub | Версия по умолчанию |
|--------|---------------|---------------------|
| Selenium-тесты Chrome | `chrome` | `148.0` |
| Selenium-тесты Firefox | `firefox` | `150.0` |
| Selenium-тесты Edge | `msedge` | `145.0` |
| `@playwright/test` в CI | `playwright-chromium` | `1.61.1` |
| Playwright + Firefox | `playwright-firefox` | `1.61.1` |
| Playwright + Safari-движок | `playwright-webkit` | `1.61.1` |
