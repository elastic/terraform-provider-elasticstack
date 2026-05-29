## 1. Shared `writeonlyhash` helper (only if not already present)

> Skip section 1 entirely if `internal/utils/writeonlyhash/` already exists (e.g. PR #3415 merged first). If skipping, jump to section 2 and adopt the helper as-is.

- [x] 1.1 Create `internal/utils/writeonlyhash/` package with `Hasher` struct, `New(resourceTypeName string) *Hasher`, `(*Hasher).Compute(value string) ([]byte, error)`, `(*Hasher).Matches(value string, storedHash []byte) bool`, `(*Hasher).PrivateStateKey(attributePath string) string`
- [x] 1.2 Implement bcrypt-backed `Compute` with per-resource-type salt derivation and configurable cost (default 10)
- [x] 1.3 Add table-driven unit tests for: roundtrip, mismatch, salt isolation across resource types, stable `PrivateStateKey` derivation, error messages do not leak input values
- [x] 1.4 Add per-method Godoc on all exported symbols matching the contract in `writeonly-secret-hashing-docs/spec.md`

## 2. Documentation for `writeonlyhash` (always; this change owns docs)

- [x] 2.1 Create `dev-docs/high-level/writeonly-secret-hashing.md` covering: why over `_wo_version`, threat model, `ModifyPlan` contract, private-state key convention (`secret_hash:<attribute_path>`), post-import behaviour, worked adoption example, anti-patterns
- [x] 2.2 Link to the helper package Godoc and to `random_password.bcrypt_hash` precedent
- [x] 2.3 Add a "Write-only secret attributes" sub-heading to `dev-docs/high-level/coding-standards.md` referencing the new doc
- [x] 2.4 Verify the worked example compiles and matches the helper's actual exported API (build/test the example as part of the helper's test suite if helper is shipped here)

## 3. Client wrappers in `internal/clients/elasticsearch/connector.go`

- [x] 3.1 Add `CreateConnector(ctx, client, plan) (id string, diags)` wrapping `POST /_connector` (no id) and `PUT /_connector/{id}` (id supplied)
- [x] 3.2 Add `GetConnector(ctx, client, id) (*get.Response, diags)` wrapping `GET /_connector/{id}`; return `nil, nil` on 404
- [x] 3.3 Add `DeleteConnector(ctx, client, id) diags` wrapping `DELETE /_connector/{id}`; tolerate 404
- [x] 3.4 Add partial-update wrappers for `_name`, `_index_name`, `_service_type`, `_native`, `_pipeline`, `_scheduling`, `_features`, `_api_key_id`, `_configuration`
- [x] 3.5 Add `CreateSyncJob(ctx, client, body) (id string, diags)` wrapping `POST /_connector/_sync_job`
- [x] 3.6 Add `GetSyncJob(ctx, client, id) (*syncjobget.Response, diags)` for the action's poll loop

## 4. Resource — package skeleton & schema

- [x] 4.1 Create `internal/elasticsearch/connector/` directory and `resource.go` registering `entitycore.NewElasticsearchResource[ContentConnectorData]` per Decision 2
- [x] 4.2 Create `models.go` with `ContentConnectorData`, branch-typed `ConfigurationValueModel`, and per-aspect nested models (`PipelineModel`, `SchedulingModel`, `FeaturesModel`, etc.)
- [x] 4.3 Create `schema.go` with the full resource schema per REQ-004 through REQ-008; mark `secret_value` as `WriteOnly + Sensitive`; gate via `entitycore.VersionRequirement` at 8.12.0 per REQ-013
- [x] 4.4 Add per-element `ObjectValidator` enforcing exactly-one-of `{string, number, bool, json, secret_value}` in `configuration_values`
- [x] 4.5 Embed `resource-description.md` for the long-form resource description; document the connector-service-first ordering and the post-import secret-baseline behaviour
- [x] 4.6 Register the resource in `provider/`

## 5. Resource — CRUD + drift detection

- [ ] 5.1 `create.go`: fan-out per Decision 4 (POST/PUT envelope → fan-out sub-updates → GET refresh)
- [ ] 5.2 `update.go`: same fan-out as Create plus `_name`/`_index_name`/`_service_type`/`_native` partials when those fields changed
- [ ] 5.3 `read.go`: implement REQ-008 read-path branch selection (prior-state branch wins; new keys use registered-schema `type`; sensitive fields skipped on import; warning diag when a sensitive API field is set in a non-secret branch)
- [ ] 5.4 `delete.go`: call `DELETE /_connector/{id}`; tolerate 404; clear the `secret_hash:configuration_values["<key>"].secret_value` private-state entry for each `configuration_values` element that existed in prior state
- [ ] 5.5 `modify_plan.go`: implement REQ-011 — hash each `secret_value` from `req.Config`, compare to stored hash, mark for update + warn on mismatch, clear on removal
- [ ] 5.6 `create.go`/`update.go`: after successful API write, store bcrypt hash for every applied `secret_value`
- [ ] 5.7 Implement REQ-009 pre-flight: if `configuration_values` is set in plan but `GET /_connector/{id}.configuration == {}`, return the schema-not-yet-registered diagnostic
- [ ] 5.8 `import.go` (or in `resource.go`): accept bare connector ID and composite form per REQ-012

## 6. Data source

- [ ] 6.1 Create `data_source.go` using `entitycore.NewElasticsearchDataSource` (envelope generic, simple read pipeline)
- [ ] 6.2 Create `data_source_schema.go` exposing every read-time attribute per REQ-010 (envelope + pipeline + scheduling + features + the full `configuration` document + all runtime telemetry: `status`, `last_seen`, `last_synced`, `last_*_sync_*` family, `filtering`, `custom_scheduling`, `sync_cursor`, `sync_now`, etc.); all `Computed`
- [ ] 6.3 Register the data source in `provider/`

## 7. Action — `connector_sync_job_create`

- [ ] 7.1 Create `internal/elasticsearch/connector/sync_job_create/` mirroring `internal/elasticsearch/cluster/snapshot_create/` layout (`action.go`, `model.go`, `schema.go`, `acc_test.go`)
- [ ] 7.2 Implement `schema.go` per REQ-SYNC-001 schema (connector_id, job_type, trigger_method, wait_for_completion, timeouts.invoke)
- [ ] 7.3 Implement `action.go`: build `POST /_connector/_sync_job` body, call API, capture sync job id; if `wait_for_completion`, poll `GET /_connector/_sync_job/{id}` every 5s until terminal status or timeout
- [ ] 7.4 Surface terminal-status `error`/`cancelled`/`suspended` as diagnostics per REQ-SYNC-001-E/F; surface timeout-exceeded per REQ-SYNC-001-D
- [ ] 7.5 Gate via `entitycore.VersionRequirement` at 8.12.0 per REQ-SYNC-002
- [ ] 7.6 Register the action in `provider/` via `ProviderWithActions`

## 8. Acceptance tests — resource

- [ ] 8.1 Create + Read + Delete with `service_type = "postgresql"`, narrow envelope only (no pipeline/scheduling/features/configuration_values)
- [ ] 8.2 Create with full envelope + pipeline + scheduling + features (no configuration_values); verify each sub-update endpoint was called via state assertions
- [ ] 8.3 Update path: change `name`/`description`/`index_name`/`pipeline`/`scheduling`/`features` and verify each triggers the correct partial-update endpoint
- [ ] 8.4 `configuration_values` round-trip per branch: `string`, `number`, `bool`, `json`; verify state matches config and subsequent plan is clean
- [ ] 8.5 `configuration_values` `secret_value` write: verify the value is sent to the API; verify state does not contain the value; verify a subsequent plan with the same value is clean; verify changing the value triggers update with a warning diagnostic; verify removing the element clears the private-state hash
- [ ] 8.6 Validator failure: `configuration_values = { x = {} }` and `configuration_values = { x = { string = "a", number = 1 } }` both fail validation
- [ ] 8.7 Sensitive-API-field warning: simulate a connector whose schema marks a field `sensitive: true`; verify the warning diagnostic is emitted when that field appears in a non-secret branch of `configuration_values`
- [ ] 8.8 Pre-flight schema-not-registered: create a connector without the connector service running, attempt `configuration_values`; verify the structured diagnostic is returned and no `_configuration` PUT was made
- [ ] 8.9 Import: bare connector ID and composite form both succeed; post-import refresh with secret in config produces no drift; first apply baselines the hash
- [ ] 8.10 Read 404: pre-delete the connector via the raw API and verify Read removes the resource from state without error
- [ ] 8.11 Delete 404: pre-delete the connector via the raw API and verify Delete succeeds
- [ ] 8.12 Version gate: attempt apply against ES < 8.12 (skip if test env can't downgrade); verify diagnostic mentions 8.12

## 9. Acceptance tests — data source

- [ ] 9.1 Read existing connector; verify all envelope, pipeline, scheduling, features, and runtime-telemetry attributes are populated
- [ ] 9.2 Read connector with non-empty `filtering` and `custom_scheduling`; verify both are exposed in state
- [ ] 9.3 Read non-existent connector; verify diagnostic error
- [ ] 9.4 Read connector whose `configuration` schema is registered with sensitive fields; verify the full schema document is exposed (sensitivity hiding is a resource-only concern; data source surfaces everything)

## 10. Acceptance tests — action

- [ ] 10.1 Async create (`wait_for_completion = false`): verify the sync job is created (visible via data source / raw API call) and the action returns immediately
- [ ] 10.2 Sync create with completion: requires a running connector service. Verify the action polls and returns once terminal status is reached. (Skip in CI if a running connector service isn't part of the test env; gate behind a build tag.)
- [ ] 10.3 Timeout exceeded: `wait_for_completion = true` with `timeouts.invoke = "5s"`; verify the timeout diagnostic
- [ ] 10.4 Error status surfaced: simulate an error-status sync job (e.g. via misconfigured connector); verify the diagnostic includes the API error
- [ ] 10.5 Connector not found: invoke with bogus `connector_id`; verify diagnostic
- [ ] 10.6 Sync job history preserved: after action completes, verify the sync job document still exists via raw API call

## 11. Examples & generated docs

- [ ] 11.1 `examples/resources/elasticstack_elasticsearch_connector/` with idiomatic Terraform showing self-managed PostgreSQL setup (envelope only; `configuration_values` requires a running service so demonstrate via a comment)
- [ ] 11.2 `examples/resources/elasticstack_elasticsearch_connector/import.sh` with `terraform import` invocation
- [ ] 11.3 `examples/data-sources/elasticstack_elasticsearch_connector/`
- [ ] 11.4 `examples/actions/elasticstack_elasticsearch_connector_sync_job_create/` (or wherever action examples live; mirror snapshot_create's location)
- [ ] 11.5 Run `make docs-generate`; verify resource, data source, and action docs render correctly
- [ ] 11.6 Run `TF_ACC=1 go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1` to validate every example plans cleanly
