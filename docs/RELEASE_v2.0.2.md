# Release v2.0.2 — qa-guru/selenoid

**Дата:** 25 июня 2026  
**Предыдущий:** [v2.0.1](RELEASE_v2.0.1.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.2

Патч-релиз: исправлено удаление ручных Playwright-сессий из UI и стабильность старта контейнеров на хосте.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **DELETE Playwright session** | `DELETE /wd/hub/session/{id}` для Playwright-сессий завершает контейнер локально, без проксирования на WS-бэкенд (исправляет 500 в UI) |
| **Docker inspect** | На хосте (не in-docker hub) контейнер проверяется сразу после старта, без ожидания порта Selenium |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.2/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 10
```
