# Release note — browsers catalog v2.3.0

**Дата:** 8 июля 2026  
**Область:** `config/browsers.json` (+ копии в cm / prod deploy)  
**Hub binary:** без смены — текущий prod binary (`v2.1.1` / последний deploy tag)

Каталог browser-образов: полный WebDriver на `qaguru/*`.

| Изменение | Описание |
|-----------|----------|
| **chrome** | default `149.0` (warm + min, 148.*) — как с v2.2.0 |
| **firefox** | Twilio → `qaguru/webdriver-firefox` · default `151.0` · path `/` |
| **msedge** | Twilio → `qaguru/webdriver-msedge` · default `145.0` · path `/` |
| **Playwright** | без изменений · `1.61.1` |

Images: [browser-image](https://github.com/qa-guru/browser-image/releases) (`webdriver/chrome-*`, `webdriver/firefox-*`, `webdriver/msedge-*`).

Prod: [selenoid.autotests.cloud](https://github.com/qa-guru/selenoid.autotests.cloud) `deploy/RELEASE_v2.3.0.md`.
