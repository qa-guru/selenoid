# Release v2.0.6 — qa-guru/selenoid

**Дата:** 26 июня 2026  
**Предыдущий:** [v2.0.5](https://github.com/qa-guru/selenoid/releases/tag/v2.0.5)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.6

Per-browser Playwright-образы вместо монолитного `qaguru/playwright` с `PW_BROWSER`.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Playwright browser keys** | `playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` |
| **Docker images** | `qaguru/playwright-chromium:1.61.1`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` |
| **Hub** | Убрана логика `PW_BROWSER` env — браузер задаётся именем в `browsers.json` и образом |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.6/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
for img in playwright-chromium playwright-firefox playwright-webkit playwright-chrome playwright-msedge; do
  docker pull "qaguru/${img}:1.61.1"
done
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 10
```

На **selenoid.autotests.cloud**:

```bash
SELENOID_VERSION=v2.0.6 ./deploy/deploy.sh
```
