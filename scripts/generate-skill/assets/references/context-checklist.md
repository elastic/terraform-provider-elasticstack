# Pre-flight context checklist

Capture these before writing HCL. State assumptions explicitly when the user hasn't said.

## Runtime and version

- [ ] Minimum Terraform version?
- [ ] Target `elastic/elasticstack` provider version (pin with `version = "~> X.Y"` or exact).
- [ ] Is `elastic/ec` (Elastic Cloud) also in play? It's a separate provider.

## Stack flavor

Pick one — behavior differs:

- Elastic Cloud (Elasticsearch Service / ESS) hosted deployment
- Elastic Cloud Serverless (many stack APIs are unavailable or differently shaped)
- Elastic Cloud on Kubernetes (ECK)
- Self-hosted / on-prem Elastic Stack

If the user hasn't said, ask — or state an assumption (default: hosted ESS, current major).

## Subsystems

- [ ] Elasticsearch — always required to configure the `elasticsearch {}` provider block.
- [ ] Kibana resources? Configure `kibana {}`.
- [ ] Fleet resources? Configure `fleet {}` (Fleet server URL + service token).

## State backend and delivery

- [ ] Where does state live? Remote backend or local?
- [ ] Is there CI enforcing plan/apply separation?
- [ ] Anything production-critical in scope? If so, require a reviewed plan before apply.

## Authentication

Prefer environment variables over inline credentials. See `references/provider.md`.
