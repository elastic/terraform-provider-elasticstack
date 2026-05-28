# Guide screenshot regeneration

Regenerate guide screenshots after dashboard or Kibana UI changes. Each guide has a Playwright script that logs into Kibana, navigates to the dashboard created by the corresponding Terraform example, and writes PNGs under `docs/guides/images/`.

## Prerequisites

- Node.js 20+
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
node guide1.mjs
node guide2.mjs
node guide3.mjs
```

## Output

Screenshots are written to `docs/guides/images/` (for example `g1-01-shell.png`, `g2-01-full.png`). Commit the PNGs to the repository after regenerating them.
