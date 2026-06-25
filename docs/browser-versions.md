# Таблица совместимости версий браузеров

Актуально для [`config/browsers.json`](../config/browsers.json). Последняя проверка: июнь 2026.

В Selenoid используются **два независимых стека**. Версии в `browsers.json` для них означают разное:

| Стек | Браузеры в `browsers.json` | Протокол | Что означает поле `version` |
|------|----------------------------|----------|------------------------------|
| **WebDriver** | `chrome`, `firefox` | Selenium (`/wd/hub`) | Мажорная версия браузера (`148.0` → Chrome 148.x) |
| **Playwright** | `chromium`, `firefox-playwright`, `webkit` | WebSocket (`/playwright/...`) | Версия npm-пакета `@playwright/test` (`1.61.1`) |

WebDriver-образы берутся из [twilio/selenoid-images](https://github.com/twilio/selenoid-images) — поддерживаемый форк Aerokube. Playwright — из [`qaguru/playwright`](https://hub.docker.com/r/qaguru/playwright) ([исходники](https://github.com/qa-guru/playwright-image), поверх `mcr.microsoft.com/playwright`).

---

## Сводная таблица (`browsers.json`)

| Имя в hub | Default | Версии в конфиге | Docker-образ | Протокол |
|-----------|---------|------------------|--------------|----------|
| `chrome` | `148.0` | 148.0, 147.0, 146.0 | `twilio/selenoid:chrome_stable_<N>` | WebDriver |
| `firefox` | `150.0` | 150.0, 149.0, 148.0 | `twilio/selenoid:firefox_stable_<N>` | WebDriver |
| `chromium` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright:v<версия>-noble` | Playwright |
| `firefox-playwright` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright:v<версия>-noble` + `PW_BROWSER=firefox` | Playwright |
| `webkit` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright:v<версия>-noble` + `PW_BROWSER=webkit` | Playwright |

---

## WebDriver: Chrome и Firefox (Twilio)

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

**Клиент:** любой Selenium WebDriver. Версия Selenium-клиента не привязана к версии браузера.

> Устаревшие образы `selenoid/chrome` и `selenoid/firefox` (Aerokube, archived) больше не используются — максимум там Chrome 128 / Firefox 125.

---

## Playwright: движки внутри каждой версии

Данные из официального [`browsers.json`](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json) Microsoft Playwright.

### Chromium (`chromium`)

| Playwright в hub | `playwrightVersion` | Chromium в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|------------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **149.0.7827.55** (rev 1228) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/chromium/1.61.1` |
| 1.61.0 | `1.61.0` | **149.0.7827.55** (rev 1228) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/chromium/1.61.0` |
| 1.60.0 | `1.60.0` | **148.0.7778.96** (rev 1223) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/chromium/1.60.0` |

### Firefox Playwright (`firefox-playwright`)

| Playwright в hub | `playwrightVersion` | Firefox в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|----------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **151.0** (rev 1532) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/firefox-playwright/1.61.1` |
| 1.61.0 | `1.61.0` | **151.0** (rev 1532) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/firefox-playwright/1.61.0` |
| 1.60.0 | `1.60.0` | **150.0.2** (rev 1522) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/firefox-playwright/1.60.0` |

### WebKit (`webkit`)

| Playwright в hub | `playwrightVersion` | WebKit в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|---------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **26.5** (rev 2311) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/webkit/1.61.1` |
| 1.61.0 | `1.61.0` | **26.5** (rev 2311) | `@playwright/test@1.61.0` | `ws://host:4444/playwright/webkit/1.61.0` |
| 1.60.0 | `1.60.0` | **26.4** (rev 2287) | `@playwright/test@1.60.0` | `ws://host:4444/playwright/webkit/1.60.0` |

---

## Правила совместимости клиента

| Что проверять | Правило | Последствие при нарушении |
|---------------|---------|---------------------------|
| Playwright npm ↔ hub | `@playwright/test` **должен совпадать** с `playwrightVersion` и версией в URL | Ошибки протокола, несовместимость CDP/BiDi |
| Playwright ↔ WebDriver | **Нельзя** подставить `chrome:148.0` вместо `chromium:1.61.1` | Сессия не создастся |
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
| Playwright 1.61.1 | [playwright-image](https://github.com/qa-guru/playwright-image) `./scripts/build.sh v1.61.1-noble` или `docker pull qaguru/playwright:v1.61.1-noble` |
| Playwright 1.61.0 | `./scripts/build.sh v1.61.0-noble` или `docker pull qaguru/playwright:v1.61.0-noble` |
| Playwright 1.60.0 | `./scripts/build.sh v1.60.0-noble` или `docker pull qaguru/playwright:v1.60.0-noble` |
| Видеозапись | `docker pull selenoid/video-recorder:latest-release` |

Базовый слой Playwright: `mcr.microsoft.com/playwright:v<версия>-noble`. Образ [`qaguru/playwright`](https://github.com/qa-guru/playwright-image) добавляет VNC поверх него.

Полный список тегов Twilio: [twilio/selenoid tags](https://hub.docker.com/r/twilio/selenoid/tags).

---

## Быстрый выбор стека

| Задача | Браузер в hub | Версия по умолчанию |
|--------|---------------|---------------------|
| Selenium-тесты Chrome | `chrome` | `148.0` |
| Selenium-тесты Firefox | `firefox` | `150.0` |
| `@playwright/test` в CI | `chromium` | `1.61.1` |
| Playwright + Firefox | `firefox-playwright` | `1.61.1` |
| Playwright + Safari-движок | `webkit` | `1.61.1` |
