# Guide screenshot regeneration

Regenerate guide screenshots after dashboard or Kibana UI changes. Each guide has a Playwright script that logs into Kibana, navigates to the dashboard created by the corresponding Terraform example, and writes PNGs under `templates/guides/images/`.

## Prerequisites

- Node.js 24.x (matches repo `make setup`)
- Dependencies installed: run the [Setup](#setup) steps below (`npm install` and `npx playwright install chromium`)
- A running Kibana 9.4+ instance with the **logs** and **eCommerce** sample datasets installed
- `terraform apply` already run for the relevant guide config under `examples/guides/`

## Setup

```bash
cd scripts/screenshots
npm install
npx playwright install chromium
```

Or use the convenience script:

```bash
npm run install-browsers
```

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `KIBANA_URL` | `http://localhost:5601` | Kibana base URL |
| `KIBANA_USER` | `elastic` | Login username |
| `KIBANA_PASS` | `password` | Login password |

## Example invocations

```bash
node scripts/screenshots/guide1.mjs
node scripts/screenshots/guide2.mjs
node scripts/screenshots/guide3.mjs
```

## Where output goes

Playwright scripts write PNGs to `templates/guides/images/` (for example `g1-01-shell.png`, `g2-01-full.png`). Running `make docs-generate` copies them to `docs/guides/images/` for publication. Commit PNGs in both directories after regenerating; `make check-docs` ensures the published copy stays in sync.
