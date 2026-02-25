# Dashboard spec bump (2026-02-25)

## Status
done

## Context
- **Problem**: `generated/kbapi/*` was regenerated from an updated Kibana dashboard API spec, causing compilation errors and drift between the dashboard resource implementation and the Terraform schema (including renamed API fields). We need the dashboard resource to match the updated API/implementation and keep pace with Kibana development.
- **Worktree**: bump-dashboards
- **Scope**: Fix **all compilation errors** introduced/exposed by the spec bump; update dashboard Terraform schema to match renamed API fields; ensure dashboard acceptance tests run at full scope (no new skips/commenting/reduction).
- **Constraints**:
  - Prefer breaking schema changes over backward-compatibility cruft (API is unreleased).
  - Do not skip/comment-out/reduce test scope to get green.
  - Use default local acceptance env vars (see Acceptance commands below).
- **Repo touchpoints**:
  - Generated client/spec: `generated/kbapi/dashboards.yaml`, `generated/kbapi/kibana.gen.go`, `generated/kbapi/transform_schema.go`
  - Dashboard resource: `internal/kibana/dashboard/{resource.go,schema.go,read.go,create.go,update.go,delete.go}`
  - Dashboard models: `internal/kibana/dashboard/models_*.go`
  - Dashboard acceptance tests + configs: `internal/kibana/dashboard/acc_*_test.go`, `internal/kibana/dashboard/testdata/**`
  - Examples (if schema fields rename): `examples/resources/elasticstack_kibana_dashboard/**`
- **Formats impacted**: none
- **Definition of done**:
  - `go test ./...` passes.
  - Dashboard acceptance tests pass with no scope reduction:
    - `TF_ACC=1 ... go test -v ./internal/kibana/dashboard -run '^TestAcc' ...` passes.
  - Any tests in changed packages pass (unit + acceptance where applicable).
  - Lint/format/docs generation checks pass (per final task commands).

## Tasks
- [x] 1) Fix `DatatableESQL` model break from kbapi regen (owner: agent)
  - **Change**: Update dashboard datatable ESQL model mappings to handle spec-changed types (e.g. pointer-to-slice fields like `Metrics`), ensuring `fromAPI`/`toAPI` correctly nil-check/deref and assign pointers where required.
  - **Files**:
    - `internal/kibana/dashboard/models_datatable_panel.go`
    - `internal/kibana/dashboard/models_datatable_panel_test.go` (if unit expectations need updating)
  - **Acceptance**:
    - `go test ./internal/kibana/dashboard -run TestNonExistent -count=0` (compile-only signal)
    - `go test ./internal/kibana/dashboard -count=1`
  - **Spec update**: mark done + append any discovered kbapi type changes/gotchas under “Additional Context”

- [x] 2) Eliminate remaining compilation errors from the kbapi spec bump (owner: agent)
  - **Change**: Run `make build`, fix the *next* compilation error(s) caused by the regenerated kbapi types; repeat within this task until `make build` compiles and runs successfully. If the remaining errors are too numerous for one session, split this task into smaller follow-up tasks in the spec (by package/area) and complete only the first split task in the current session.
  - **Files**: any packages failing to compile (expected touchpoints include `internal/kibana/**`, `internal/clients/**`, and any packages referencing regenerated kbapi structs/enums)
  - **Acceptance**:
    - `go test ./... -count=1`
  - **Spec update**: mark done (or split tasks + document remaining failing packages/errors)

- [x] 3) Rename Terraform dashboard schema fields to match renamed API fields (breaking change) (owner: agent)
  - **Change**: For any API field renames introduced by the updated dashboard spec/client, rename the corresponding Terraform schema attributes and model fields to match the new API names (no BWC shims). Update read/create/update mappings accordingly, and update any acceptance test configs and examples that reference renamed attributes.
  - **Files**:
    - `internal/kibana/dashboard/schema.go`
    - `internal/kibana/dashboard/models*.go` (as needed)
    - `internal/kibana/dashboard/testdata/**/main.tf` (as needed)
    - `examples/resources/elasticstack_kibana_dashboard/**` (as needed)
  - **Acceptance**:
    - `make fmt` (ensures `terraform fmt --recursive` runs too)
    - `go test ./internal/kibana/dashboard -count=1`
  - **Spec update**: mark done + record the exact attribute renames performed (old -> new) in “Additional Context”

- [x] 4) Run full dashboard acceptance test suite with default local auth (no skips) (owner: agent)
  - **Change**: Execute dashboard acceptance tests directly with `go test` against the already-running stack, using the repo’s default env var auth values. Fix any acceptance failures by updating resource/schema/mappings/testdata (without skipping/commenting/reducing scope).
  - **Files**:
    - `internal/kibana/dashboard/acc_*_test.go`
    - `internal/kibana/dashboard/testdata/**`
    - `internal/kibana/dashboard/*.go` (resource + models as needed)
  - **Acceptance** (use explicit env vars):
    - `ELASTICSEARCH_ENDPOINTS=http://localhost:9200 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=password KIBANA_ENDPOINT=http://localhost:5601 TF_ACC=1 go test -v ./internal/kibana/dashboard -run '^TestAcc' -count=1 -timeout 120m`
  - **Spec update**: mark done + note any Kibana-version/spec quirks encountered

- [x] 5) Run all checks and fix issues (owner: agent)
  - **Change**: Run repo formatting, linting, docs generation, and tests; fix any failures found (no skipping).
  - **Files**: any files with issues
  - **Acceptance**:
    - `make fmt`
    - `make lint`
    - `go test ./... -count=1`
    - `ELASTICSEARCH_ENDPOINTS=http://localhost:9200 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=password KIBANA_ENDPOINT=http://localhost:5601 TF_ACC=1 go test -v ./internal/kibana/dashboard -run '^TestAcc' -count=1 -timeout 120m`
  - **Spec update**: mark done

- [x] 6) Create commit (owner: agent)
  - **Change**: Stage all changes and create a descriptive commit (include schema rename + spec alignment in message).
  - **Files**: none (git operation)
  - **Acceptance**: `git status` shows clean working tree; `git log -1` shows the new commit
  - **Spec update**: mark done; set `## Status` to `done` and update spec index row status/date/summary

- [x] 7) Split root level `query.query` into `query.text` and `query.json` (owner: agent)
  - **Change**:
    - Replace the single root level `query.query` attribute with two attributes:
      - `query.text` (string): for text-based queries (e.g. KQL/Lucene).
      - `query.json` (string): for structured query JSON (persisted as a JSON-encoded string).
    - Only the `query.query` attribute in `getSchema` should be modifier. Other panel specific queries should be left untouched 
    - Enforce **mutual exclusivity**: exactly one of `query.text` / `query.json` may be set (and at least one must be set when the `query` block is present).
    - Persist state on read/GET based on the **type returned by the API**:
      - If the API returns a string query, populate `query.text` and ensure `query.json` is null/unknown.
      - If the API returns a structured JSON object, populate `query.json` (via `jsonencode`) and ensure `query.text` is null/unknown.
    - Update create/update mappings and acceptance test configs accordingly.
  - **Files**:
    - `internal/kibana/dashboard/schema.go`
    - `internal/kibana/dashboard/models_*.go` (query/searchSource mapping)
    - `internal/kibana/dashboard/read.go` (if special-casing is needed on read)
    - `internal/kibana/dashboard/testdata/**/main.tf` + `internal/kibana/dashboard/acc_*_test.go` (as needed)
    - `examples/resources/elasticstack_kibana_dashboard/**` (as needed)
  - **Acceptance**:
    - `go test ./... -count=1`
    - `ELASTICSEARCH_ENDPOINTS=http://localhost:9200 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=password KIBANA_ENDPOINT=http://localhost:5601 TF_ACC=1 go test -v ./internal/kibana/dashboard -run '^TestAcc' -count=1 -timeout 120m`
  - **Spec update**: record final schema shape + any import/read preservation quirks discovered

- [x] 8) Create follow-up commit for Task 7 (owner: agent)
  - **Change**: Stage all changes from Task 7 and create a descriptive commit message focused on the query schema split and read preservation.
  - **Files**: none (git operation)
  - **Acceptance**: `git status` shows clean working tree; `git log -1` shows the new commit
  - **Spec update**: mark done; set `## Status` to `done` and update spec index row status/date/summary

## Additional Context
- Default acceptance env vars (per `.github/copilot-instructions.md`):
  - `ELASTICSEARCH_ENDPOINTS=http://localhost:9200`
  - `ELASTICSEARCH_USERNAME=elastic`
  - `ELASTICSEARCH_PASSWORD=password`
  - `KIBANA_ENDPOINT=http://localhost:5601`
  - `TF_ACC=1`
- Known first compilation break after kbapi regen: `internal/kibana/dashboard/models_datatable_panel.go` treats `api.Metrics` like a slice, but kbapi now defines it as a pointer-to-slice (`*[]...`) for `DatatableESQL`.
- Gotcha: `kbapi.DatatableESQL.Metrics` is now `*[]kbapi.DatatableESQLMetric` (while `kbapi.DatatableNoESQL.Metrics` remains a required `[]...`), so model `fromAPI` must nil-check/deref and `toAPI` must assign a pointer.
- Gotcha: `go test ./internal/kibana/dashboard -count=1` will run acceptance tests if `TF_ACC` is set in your shell environment; ensure the variable is not set (e.g. `unset TF_ACC`) when you intend a unit-only run.
- Discovery: After completing Task 1, `make build` and `TF_ACC= go test ./... -count=1` are already clean on `bump-dashboards` (no remaining kbapi-regeneration compilation errors to chase at this stage).
- Task 3 schema renames (breaking; old -> new):
  - `time_from` -> `time_range.from`
  - `time_to` -> `time_range.to`
  - `time_range_mode` -> `time_range.mode`
  - `refresh_interval_pause` -> `refresh_interval.pause`
  - `refresh_interval_value` -> `refresh_interval.value`
  - `query_language` -> `query.language`
  - `query.query` -> split into:
    - `query.text` (plain query string)
    - `query.json` (JSON-encoded object string)

- Import/read model gotcha: Terraform Plugin Framework decoding during import will present required nested blocks as `null` (except `id`), so `dashboardModel` must use pointer fields for `time_range`, `refresh_interval`, and `query` to avoid conversion errors.
- `time_range.mode` is not returned by the API on GET; the provider preserves a previously-known value to avoid post-apply inconsistencies.
- Query split (Task 7):
  - **State mapping**: on read, the provider sets exactly one of `query.text` / `query.json` based on whether the Kibana API returns a string or an object for `data.query.query`.
  - **Validation gotcha**: `objectvalidator.ExactlyOneOf()` is incompatible with a query block that also has `language` set; use `objectvalidator.AtLeastOneOf(text,json)` plus `ConflictsWith` validators on `text`/`json` to enforce “exactly one”.
- Task 7 verification (2026-02-25):
  - `make fmt`, `make lint`, `go test ./... -count=1`, and the full dashboard acc suite all passed after migrating configs/examples to `query.text`/`query.json`.
- Metric chart `metrics[*].config_json` drift: Kibana may inject defaults (e.g. `fit`, `empty_as_null`, `alignments.value`, `icon.align`, `label_position`) that can cause post-apply inconsistencies and import-verify diffs. The provider strips known defaults on read to keep state stable, while still using semantic equality that ignores missing defaults.
- Task 5 lint follow-ups:
  - `internal/kibana/dashboard/schema.go`: introduced constants for repeated literals (`"right"`, `"number"`, `"percent"`) to satisfy `goconst`.
  - `internal/clients/kibanaoapi/dashboards.go`: renamed `Id` -> `ID` in the request-body flattener wrapper struct to satisfy `revive`.
- Local-only agent artifacts are ignored via `.gitignore`: `.ralph/` and `spec/sessions/` (prompt/run JSONL logs).
