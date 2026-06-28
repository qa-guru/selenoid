# Release v2.0.3 — qa-guru/selenoid

**Дата:** 25 июня 2026  
**Предыдущий:** [v2.0.2](RELEASE_v2.0.2.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.3

Патч — исправлен чёрный экран VNC в ручных Playwright-сессиях.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Playwright VNC (manual)** | Headed-браузер стартует сразу в фоне, без `sleep 2` и без привязки к `run-server` |
| **Playwright-образы** | per-browser: `qaguru/playwright-chromium`, `playwright-firefox`, …; `launch-headed-browser.js` в `/opt/playwright/` |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.3/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
docker pull qaguru/playwright-chromium:1.61.1
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 10
```
