# Release v2.1.8 — qa-guru/selenoid

**Дата:** 8 июля 2026  
**Предыдущий:** [v2.1.7](RELEASE_v2.1.7.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.8

Патч CI/toolchain: **Go 1.26.5** для чистого blocking `govulncheck` (GO-2026-5856).

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **toolchain** | `go1.26.4` → **go1.26.5** в `go.mod` и `ci/test.sh` |
| **govulncheck** | CI снова зелёный на `main` |

Функциональных изменений hub нет — только toolchain bump.

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.8/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.1.8`

Prod: deploy workflow с `version=v2.1.8`.
