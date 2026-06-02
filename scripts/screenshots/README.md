# Guide screenshot regeneration

Regenerate guide screenshots after dashboard or Kibana UI changes. Each guide has a Playwright script that logs into Kibana, navigates to the dashboard created by the corresponding Terraform example, and writes PNGs under `templates/guides/images/`.

## Prerequisites

- Node.js 24.x (matches repo `make setup`)
- Dependencies installed: run the [Setup](#setup) steps below (`npm install` and `npx playwright install chromium`)
- A running Kibana 9.4+ instance with the **logs** and **eCommerce** sample datasets installed
- `terraform apply` already run for the relevant guide config under `examples/guides/` (creates the panel-populated dashboards each script screenshots)

> Note: `guide1.mjs` also creates a short-lived **empty** dashboard via the Kibana REST API to capture `g1-01-shell.png`, then deletes it. No extra manual setup is required.

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
| `SCREENSHOT_ONLY` | _(unset)_ | Optional. Comma-separated PNG filenames to capture (e.g. `g2-01-full.png,g2-02-filtered.png`). Re-runs only those shots without redoing the rest. |
| `DASHBOARD_ID` | _(unset)_ | Optional. Kibana dashboard saved-object ID. Skips title-based lookup when multiple dashboards share the same title. Useful for targeted re-captures. |

## Example invocations

```bash
node scripts/screenshots/guide1.mjs
node scripts/screenshots/guide2.mjs
node scripts/screenshots/guide3.mjs
```

## Where output goes

Scripts write to `templates/guides/images/` — that directory is the source of truth; commit regenerated PNGs there. `make docs-generate` copies them to `docs/guides/images/` for publication; `make check-docs` verifies the published copy is in sync. Contributors normally do not commit `docs/guides/images/` by hand (it is generated). CI still expects `docs/guides/images/` to match `templates/guides/images/` after docs generation (task 8.2 committed the initial published copy; re-run `make docs-generate` when adding or replacing images).
