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

### Examples PlanOnly harness (`TestAccExamples_planOnly`)

`*.tf` files under `examples/resources/` and `examples/data-sources/` (except harness skip-lists) are planned in isolation by `TestAccExamples_planOnly` in `internal/acctest/`. For contributor expectations—self-contained modules, acceptance environment—see the **Example snippets** section in [`development-workflow.md`](./development-workflow.md).

```bash
TF_ACC=1 go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1
```

### Required environment variables (common)

Targeted acceptance test runs commonly require the following environment variables:

- `ELASTICSEARCH_ENDPOINTS` (default: `http://localhost:9200`)
- `ELASTICSEARCH_USERNAME` (default: `elastic`)
- `ELASTICSEARCH_PASSWORD` (default: `password`)
- `KIBANA_ENDPOINT` (default: `http://localhost:5601`)
- `TF_ACC=1`

Example targeted run:

```bash
ELASTICSEARCH_ENDPOINTS=http://localhost:9200 \
ELASTICSEARCH_USERNAME=elastic \
ELASTICSEARCH_PASSWORD=password \
KIBANA_ENDPOINT=http://localhost:5601 \
TF_ACC=1 \
go test -v -run TestAccResourceName ./path/to/package
```

### Acceptance test coverage expectations

See “Testing” in [`coding-standards.md`](./coding-standards.md).

