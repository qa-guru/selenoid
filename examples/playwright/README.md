# Пример Playwright

<!-- stack-branches-note:start -->
**Stack pin:** стабильные сборки всего стека — ветки [`selenoid2-1.45-engine26.1-go1.26-react16`](https://github.com/qa-guru/selenoid/tree/selenoid2-1.45-engine26.1-go1.26-react16) (**v2.2.1**, API 1.45 / Engine 26.1.x / Go 1.26.5 / React 16) и [`selenoid2-1.55-engine29.6-go1.26-react18`](https://github.com/qa-guru/selenoid/tree/selenoid2-1.55-engine29.6-go1.26-react18) (**v2.3.0**, API 1.55 / Engine 29.6+ / Go 1.26.5 / React 18). Детали — [`STACK-PIN.md`](../../STACK-PIN.md) · [корневой README](../../README.md) · [monorepo SSOT](https://github.com/qa-guru/zero-design-system/blob/master/projects/selenoid-home/README.md).
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
