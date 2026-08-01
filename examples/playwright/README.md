# Пример Playwright

<!-- stack-branches-note:start -->
> **Stack pin:** prod hub releases с **`main`** (Selenoid 3); pin-ветки 2.x — frozen rollback. Детали — [`STACK-PIN.md`](../../STACK-PIN.md) · [monorepo SSOT](https://github.com/qa-guru/zero-design-system/blob/master/projects/selenoid-home/README.md).
<!-- stack-branches-note:end -->


Smoke-тест нативной поддержки Playwright в Selenoid.

## Предварительные условия

```bash
go build -o selenoid .
docker pull qaguru/playwright-chromium:1.61.1
./selenoid -conf config/browsers.json -limit 5
```

## Запуск

```bash
npm install
cp .env.example .env
npm test
```

Подробности endpoint — в [docs/playwright.md](../../docs/playwright.md).
