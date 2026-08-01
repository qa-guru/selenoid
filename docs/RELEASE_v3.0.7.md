# Release v3.0.7 — qa-guru/selenoid

**Дата:** 1 августа 2026  
**Предыдущий:** [v3.0.6](https://github.com/qa-guru/selenoid/releases/tag/v3.0.6)

## Что нового

| Изменение | Описание |
|-----------|----------|
| **warmReady / warmTotal** | Hub probes warm-pool orchestrator (`-warm-pool-url` / `SELENOID_WARM_POOL_URL`) and exposes counts on flat `/status` for selenoid-ui WARM |

```bash
# box1 unit (see selenoid-warm-pool/deploy/selenoid-hub.service)
-warm-pool-url http://127.0.0.1:9090
```
