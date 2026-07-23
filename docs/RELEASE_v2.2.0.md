# Release v2.2.0 — qa-guru/selenoid

**Дата:** 10 июля 2026  
**Предыдущий:** [v2.1.8](RELEASE_v2.1.8.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.2.0 *(tag при cut)*  
**Stack cut:** hub + UI + cm → единый **v2.2.0** (фаза G `release-selenoid-stack`).

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **browsers catalog** | chrome default **149.0**; firefox/msedge → `qaguru/webdriver-*` (151 / 145); Playwright **1.61.1** без изменений |
| **video** | default recorder → `qaguru/video-recorder:latest`; запись на modern Docker networks |
| **Docker network** | reachability hub↔browser на custom network |
| **tests / CI** | Go unit subpackages; race/flake fixes; smoke `workflow_call` → selenoid-tests |
| **toolchain** | Go **1.26.5**; `DOCKER_API_VERSION=1.45` / Engine **26.1.x** |

Images (browser nodes): [browser-image](https://github.com/qa-guru/browser-image) — теги `webdriver/chrome-149`, `firefox-151`, `msedge-145` (+ `-min`); не тот же git-тег `v2.2.0`.

Prod: [selenoid.qa.guru](https://github.com/qa-guru/selenoid.qa.guru) после cut — отдельный deploy.

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.2.0/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.2.0`

Prod: deploy workflow с `version=v2.2.0`.

Связанные: [selenoid-ui v2.2.0](https://github.com/qa-guru/selenoid-ui/releases/tag/v2.2.0), [cm v2.2.0](https://github.com/qa-guru/cm/releases/tag/v2.2.0).

---

## Cut checklist (ручной)

1. Commit всех изменений на `main` (docs + код фазы D–F).
2. `git tag -a v2.2.0 -m "v2.2.0"` → `git push origin main --tags` *(только по явной команде)*.
3. GitHub Release / CI `release.yml` → assets `dist/selenoid_*`; Docker `qaguru/selenoid:v2.2.0`.
4. OUT: `warm-pool-orchestrator/`.
