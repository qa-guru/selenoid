# Release v2.0.1 — qa-guru/selenoid

**Дата:** 25 июня 2026  
**Предыдущий:** [v2.0.0](RELEASE_v2.0.0.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.1

Патч-релиз: выровнены версии **Go** и **Docker API** для предсказуемой работы hub на сервере и локально.

---

## Что нового

| Изменение | Было | Стало |
|-----------|------|-------|
| **Go** (`go.mod`, CI) | 1.22 | **1.23.x** |
| **Docker API** (hub) | не зафиксирован | **1.45** (`ENV` в образе, `scripts/start-selenoid.sh`) |
| **Docker Engine** (рекомендация) | — | **26.1.x** (совпадает с `docker/docker v26.1.5`) |

### Скрипты

| Скрипт | Назначение |
|--------|------------|
| `scripts/start-selenoid.sh` | Hub с `DOCKER_API_VERSION=1.45` |
| `scripts/build-selenoid.sh` | Сборка через Go 1.23 или `golang:1.23` |
| `scripts/check-toolchain.sh` | Проверка окружения |
| `scripts/docker-engine-pin-ubuntu.sh` | Pin Engine 26.1.x на Ubuntu |

Файлы `.go-version` / `.tool-versions` — для goenv, mise, asdf.

---

## Установка / обновление

```bash
# Бинарник
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.1/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 10

# Docker-образ
docker pull qaguru/selenoid:v2.0.1

# cm
./cm selenoid start
```

На Mac с Docker Desktop (Engine 27.x) hub работает с `DOCKER_API_VERSION=1.45` — поведение совпадает с сервером на 26.1.

---

## Связанные релизы

| Компонент | Версия |
|-----------|--------|
| [qa-guru/selenoid-ui](https://github.com/qa-guru/selenoid-ui/releases/tag/v2.0.1) | v2.0.1 |
| [qa-guru/cm](https://github.com/qa-guru/cm/releases/tag/v2.0.1) | v2.0.1 |

Функциональность Playwright / WebDriver без изменений относительно v2.0.0.
