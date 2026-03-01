# Development workflow

For full contributor guidance (setup, PR expectations), see [`CONTRIBUTING.md`](../../CONTRIBUTING.md).

## Typical change loop

- Read the problem statement and identify the affected area (Elasticsearch vs Kibana vs Fleet, etc).
- Add acceptance test cases reproducing bugs, or validating new work. 
- For bugs, run the new acceptance tests verifying that they fail as expected, i.e they reproduce the original issue. 
- The System User resource (see `internal/elasticsearch/security/system_user` referenced from [`CODING_STANDARDS.md`](../../CODING_STANDARDS.md)) is the canonical example for new resources. Follow it.
- Make small, reviewable changes.
- Keep generated artifacts up to date (docs and generated clients when applicable).
- Run the narrowest tests that prove correctness, then broaden as appropriate.

## Common make targets

The canonical list is the root `Makefile`, but the usual ones are:

- `make lint`
- `make test`
- `make testacc` (requires Docker and `TF_ACC=1`)
- `make docs-generate`
