# Release v3.0.12 — qa-guru/selenoid

**Дата:** 15 августа 2026  
**Предыдущий:** [v3.0.11](https://github.com/qa-guru/selenoid/releases/tag/v3.0.11)  
**Stack cut:** hub → **v3.0.12**; cm → **v3.0.2**; UI → **v3.0.33**.

## Что нового

| Изменение | Описание |
|-----------|----------|
| **warmSlots / hotSlots** | Hub `/status` includes orchestrator slot rows (`id`, `browser`, `protocol`, `pool`, `reservedBy`) split by `pool=hot`. Counts `warmReady`/`hotReady` unchanged. |

```bash
docker pull qaguru/selenoid:v3.0.12
```
