# Release v2.0.8 — qa-guru/selenoid

**Дата:** 26 июня 2026  
**Предыдущий:** [v2.0.6](RELEASE_v2.0.6.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.8

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **`msedge` в browsers.json** | Ключ Edge выровнен с chrome/firefox |
| **MicrosoftEdge alias** | `browserName: MicrosoftEdge` в caps → lookup `msedge` |
| **W3C caps** | В ответе сессии `browserName` нормализуется в `MicrosoftEdge` |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.8/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```
