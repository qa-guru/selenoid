# Release v2.3.0 — qa-guru/selenoid

**Дата:** 12 июля 2026  
**Предыдущий:** [v2.2.1](RELEASE_v2.2.1.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.3.0 *(tag при cut)*  
**Stack cut:** hub + UI + cm → единый **v2.3.0**.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Docker API 1.55** | Moby client `github.com/moby/moby/api v1.55.0` + `moby/client v0.5.0`; canon `DOCKER_API_VERSION=1.55` |
| **Docker Engine** | Рекомендуется **29.x** (API 1.55); breaking change vs v2.2.x (Engine 26.1.x / API 1.45) |
| **Client negotiation** | `createCompatibleDockerClient` — downgrade при отсутствии env-пина |
| **Browser SSOT** | Без изменений каталога: chrome **149.0**, firefox **151.0**, msedge **145.0**, Playwright **1.61.1** |

Browser-image теги (`playwright/1.61.1`, `webdriver/*`) — без изменений.

---

## Миграция с v2.2.x

1. Обновить Docker Engine до **29.x** (или задать `DOCKER_API_VERSION` под ваш daemon).
2. Обновить hub binary/image до `v2.3.0`.
3. Обновить UI и cm до `v2.3.0` (cm передаёт hub `DOCKER_API_VERSION=1.55`).

**Не в scope v2.3.0:** prod [selenoid.autotests.cloud](https://selenoid.autotests.cloud) — остаётся на v2.2.x (старый стек); обновление prod — отдельный deploy-cut.

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.3.0/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.55 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.3.0`

Связанные: [selenoid-ui v2.3.0](https://github.com/qa-guru/selenoid-ui/releases/tag/v2.3.0), [cm v2.3.0](https://github.com/qa-guru/cm/releases/tag/v2.3.0), [selenoid-tests](https://github.com/qa-guru/selenoid-tests).

---

## Cut checklist (ручной)

1. Commit docs + code на `main`.
2. `git tag -a v2.3.0 -m "v2.3.0"` → `git push origin main --tags` *(только по явной команде)*.
3. GitHub Release / CI `release.yml` → assets `dist/selenoid_*`; Docker `qaguru/selenoid:v2.3.0`.
4. deploy-smoke dispatch → selenoid-tests (CI/github profile; **не** prod autotests.cloud).
5. OUT: `warm-pool-orchestrator/`; prod [selenoid.autotests.cloud](https://selenoid.autotests.cloud) (v2.2.x, обновление отложено).
