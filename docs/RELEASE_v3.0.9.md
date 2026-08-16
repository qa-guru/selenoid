# Release v3.0.9 — qa-guru/selenoid

**Дата:** 13 августа 2026  
**Предыдущий:** [v3.0.8](https://github.com/qa-guru/selenoid/releases/tag/v3.0.8)  
**Stack cut:** hub → **v3.0.9**; cm → **v3.0.2**; UI → **v3.0.29**.

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Container-reuse Chrome WD** | With `-warm-pool-url`, hub `POST /pool/reserve` `{loopback:true}` and reuses a host-reachable warm Chrome instead of `docker run`. Cold fallback on 409 / non-loopback / video / VNC / HAR / wait fail. |
| **Prod pin** | Box1 `config.hub.yaml` is docker-DNS only → **409 → cold**. Metrics `warmReady`/`warmTotal` unchanged. Attach needs published loopback WD ports (not this pin). |

Operator guide: [CONTAINER-REUSE.md](CONTAINER-REUSE.md) · ADR: [ADR-container-reuse.md](ADR-container-reuse.md)

```bash
docker pull qaguru/selenoid:v3.0.9
```
