# Testing

The provider has unit tests and acceptance tests.

## Unit tests

- Run: `make test`
- For general contributor notes, see [`CONTRIBUTING.md`](../../CONTRIBUTING.md).

## Acceptance tests

Acceptance tests require a running Elastic Stack and `TF_ACC=1`. 

- Prefer running targeted tests with `go test -v [-run 'filter'] <package>`
- When instructed, run the full acceptance test suite with `make testacc`
- The Elastic stack is almost certainly already running. Check before considering starting a new environment (`curl -u $ELASTICSEARCH_USERNAME:$ELASTICSEARCH_PASSWORD $ELASTICSEARCH_ENDPOINTS`). 
- Start local stack services if required (set `STACK_VERSION` if a specific version is required):
  - `make docker-fleet`

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

See “Testing” in [`CODING_STANDARDS.md`](../../CODING_STANDARDS.md).

