# Testing

The provider has unit tests and acceptance tests.

## Unit tests

- Run: `make test`
- For general contributor notes, see [`contributing.md`](./contributing.md).

## Acceptance tests

Acceptance tests require a running Elastic Stack and `TF_ACC=1`.

General workflow (all acceptance coverage, not only the examples harness):

- Prefer targeted runs: `go test -v [-run 'filter'] <package>`
- When instructed, run the full suite with `make testacc`
- The Elastic stack may already be running; check before starting a new environment (`curl -u $ELASTICSEARCH_USERNAME:$ELASTICSEARCH_PASSWORD $ELASTICSEARCH_ENDPOINTS`)
- To start local stack services if needed (set `STACK_VERSION` when a specific version is required): `make docker-fleet`

### Worktree stack isolation and ports

When worktrunk creates a worktree, it brings up an **isolated Elastic Stack** (Elasticsearch + Kibana + Fleet) in Docker that is unique to that worktree. This avoids conflicts with the stack on `main` or in other branches.

Port numbers are deterministically randomised per branch to prevent collisions:

- **Elasticsearch**: `10000 + (branch | hash_port) % 5000`
- **Kibana**: `15000 + ((branch ~ '-kb') | hash_port) % 5000`

On `main` (or any checkout without worktrunk) ports fall back to the standard defaults (`9200` and `5601`) via the Makefile.

### Environment variables (`.env`)

In a worktrunk-created worktree, `.config/wt.toml` auto-generates `.env` with all common acceptance variables already set to the worktree's isolated ports:

- `ELASTICSEARCH_PORT`
- `KIBANA_PORT`
- `ELASTICSEARCH_ENDPOINTS`
- `ELASTICSEARCH_URL`
- `ELASTICSEARCH_USERNAME`
- `KIBANA_ENDPOINT`
- `KIBANA_USERNAME`

After exporting `.env` you can run acceptance tests directly:

```bash
source .env
TF_ACC=1 go test -v -run TestAccResourceName ./path/to/package
```

On `main` a `.env` may or may not already exist—if it does, you can `source` it like in a worktree. Otherwise, manually export the default variables before running tests:

```bash
# On main without .env
export ELASTICSEARCH_ENDPOINTS=http://localhost:9200
export ELASTICSEARCH_USERNAME=elastic
export ELASTICSEARCH_PASSWORD=password
export KIBANA_ENDPOINT=http://localhost:5601
TF_ACC=1 go test -v -run TestAccResourceName ./path/to/package
```

### Examples PlanOnly harness (`TestAccExamples_planOnly`)

`*.tf` files under `examples/resources/` and `examples/data-sources/` (except harness skip-lists) are planned in isolation by `TestAccExamples_planOnly` in `internal/acctest/`. For contributor expectations—self-contained modules, acceptance environment—see the **Example snippets** section in [`development-workflow.md`](./development-workflow.md).

```bash
source .env  # worktree only
TF_ACC=1 go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1
```

### Acceptance test coverage expectations

See "Testing" in [`coding-standards.md`](./coding-standards.md).

