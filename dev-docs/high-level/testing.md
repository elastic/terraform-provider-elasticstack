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
- To start local stack services if needed (set `STACK_VERSION` when a specific version is required; the default in `.env.template` is bumped automatically to the latest GA by the version-matrix workflow): `make docker-fleet`

### Full-suite acceptance suites in CI

Push, workflow-dispatch, and merge-queue runs keep two `general` shards for every pinned Stack version. For Stack 9.4 and later (including snapshots), one additional `dashboard` job runs the measured hotspot packages `internal/kibana/dashboard` and `internal/kibana/dashboard/panel/lensxy` on its own Stack. The general shards exclude those two packages but retain the other dashboard descendants. Older versions retain the original all-package split. Each job uses the same Stack setup, test concurrency, diagnostics, teardown, and snapshot failure policy. Provider Gate observes the aggregate acceptance job result.

Pull requests keep the existing targeted package plan and two-shard limit, with no extra dashboard job. Empty targeted plans skip Stack setup and acceptance execution.

Local `make testacc` defaults to the `all` suite, which includes dashboard packages regardless of `STACK_VERSION`. To inspect full-suite assignments without running tests:

```bash
go run ./scripts/acceptance-test-packages --suite=general --stack-version=9.4.0 --total-shards=2 --shard-index=0
go run ./scripts/acceptance-test-packages --suite=general --stack-version=9.4.0 --total-shards=2 --shard-index=1
go run ./scripts/acceptance-test-packages --suite=dashboard --stack-version=9.4.0 --total-shards=1 --shard-index=0
```

To run a selected suite against an existing Stack, use matching values, for example:

```bash
make testacc STACK_VERSION=9.4.0 ACCTEST_SUITE=dashboard ACCTEST_TOTAL_SHARDS=1 ACCTEST_SHARD_INDEX=0
```

Set `STACK_VERSION` to the server version when selecting `general`; it determines whether the two hotspot packages are excluded. The selector rejects unknown suites, nonpositive shard counts, and indices outside `[0, total_shards)` before tests execute. Local suite selection uses the existing Stack; reproduce CI isolation with separate worktree Stacks if needed.

To evaluate isolation, compare equivalent full-suite CI runs using the same provider commit, pinned Stack versions, runner types, and test concurrency. Job names expose the version, suite, and shard so dashboard elapsed time can be compared separately from each general shard. Record acceptance wall-clock duration (the critical path across parallel jobs) and the sum of all acceptance job minutes, including setup and teardown. The additional Stack can increase total runner time even if wall-clock duration falls. Compare individual dashboard test durations as well; isolation alone does not establish a performance improvement or identify the cause of slow tests.

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

### Provider-defined actions require Terraform 1.14+

Acceptance tests for provider-defined actions (e.g. `elasticstack_elasticsearch_connector_sync_job_create`, `elasticstack_elasticsearch_snapshot_create`) require Terraform Core ≥1.14, which introduced provider-defined actions. The default Terraform used by `terraform-plugin-testing` is older and will SKIP these tests silently.

Run action acceptance tests with an explicit Terraform version override:

```bash
TF_ACC=1 TF_ACC_TERRAFORM_VERSION=1.15.5 go test -v -run '^TestAccActionConnector' ./internal/elasticsearch/connector/sync_job_create/...
```

Or set `TF_ACC_TERRAFORM_PATH=/path/to/terraform` if you've installed Terraform 1.14+ system-wide.

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

