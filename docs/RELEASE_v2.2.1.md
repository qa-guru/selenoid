# Release v2.2.1 — qa-guru/selenoid

**Дата:** 10 июля 2026  
**Предыдущий:** [v2.2.0](RELEASE_v2.2.0.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.2.1 *(tag при cut)*  
**Stack cut:** hub + UI + cm → единый **v2.2.1** (patch после v2.2.0).

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Tag alignment** | `v2.2.1` = HEAD (post-cut CI: `workflow_dispatch` upload assets для release) |
| **README** | Единый блок «Экосистема qa-guru Selenoid»: все GitHub-репо + [selenoid-tests](https://github.com/qa-guru/selenoid-tests) + [Docker Hub qaguru](https://hub.docker.com/u/qaguru) |
| **Runtime** | Без изменений browsers catalog / hub logic относительно v2.2.0 |

Browser-image теги (`playwright/1.61.1`, `webdriver/*`) — без изменений.

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.2.1/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.2.1`

Prod: deploy workflow с `version=v2.2.1`.

Связанные: [selenoid-ui v2.2.1](https://github.com/qa-guru/selenoid-ui/releases/tag/v2.2.1), [cm v2.2.1](https://github.com/qa-guru/cm/releases/tag/v2.2.1), [selenoid-tests](https://github.com/qa-guru/selenoid-tests).

---

## Cut checklist (ручной)

1. Commit docs на `main`.
2. `git tag -a v2.2.1 -m "v2.2.1"` → `git push origin main --tags` *(только по явной команде)*.
3. GitHub Release / CI `release.yml` → assets `dist/selenoid_*`; Docker `qaguru/selenoid:v2.2.1`.
4. deploy-smoke dispatch → selenoid-tests.
5. OUT: `warm-pool-orchestrator/`.
