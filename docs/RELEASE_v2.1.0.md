# Release v2.1.0 — qa-guru/selenoid

**Дата:** 28 июня 2026  
**Предыдущий:** [v2.0.9](RELEASE_v2.0.9.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.0

Инфраструктурный релиз: **единая версия стека v2.1.0**, prod-деплой в [qa-guru/selenoid.qa.guru](https://github.com/qa-guru/selenoid.qa.guru). Код hub без изменений относительно v2.0.9.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Prod deploy** | [selenoid.qa.guru](https://github.com/qa-guru/selenoid.qa.guru) — отдельный репозиторий |
| **Стек v2.1.0** | cm, hub, UI, deploy — одна версия **v2.1.0** |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.0/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```
