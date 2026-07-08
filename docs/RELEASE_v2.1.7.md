# Release v2.1.7 — qa-guru/selenoid

**Дата:** 8 июля 2026  
**Предыдущий:** [v2.1.6](RELEASE_v2.1.6.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.7

Патч hub: совместимость catalog-ключей `*-min` с geckodriver / строгими драйверами.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Proxy rewrite** | Перед `POST` в browser driver суффикс `-min` снимается с `browserVersion` / `version` |
| **Зачем** | Ключ `151.0-min` выбирает образ `qaguru/webdriver-firefox:151-min`, но Firefox внутри — `151.0`; geckodriver иначе отвечает `Unable to find a matching set of capabilities` |
| **Тесты** | `TestCatalogVersionForDriver`, `TestRewriteCatalogBrowserVersionForDriver`, `TestSessionCreatedStripsMinSuffixForDriver` |

Образ / выбор версии в `browsers.json` **не меняются** — rewrite только в сторону driver.

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.7/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.1.7`

Prod: deploy workflow с `version=v2.1.7`.
