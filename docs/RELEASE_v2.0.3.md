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
| **Образ `qaguru/playwright`** | `launch-headed-browser.js` использует `browser.launch({ headless: false })` вместо нерабочего `connect()` к `run-server` |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.3/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
docker pull qaguru/playwright:v1.61.1-noble
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 10
```

На **selenoid.autotests.cloud**:

```bash
SELENOID_VERSION=v2.0.3 ./deploy/deploy.sh
```
