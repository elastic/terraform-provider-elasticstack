## 1. Shared timeouts plumbing

- [x] 1.1 Create `internal/entitycore/resource_timeouts.go` defining `ResourceTimeoutsField` (embeddable struct holding field `Timeouts timeouts.Value` tagged `tfsdk:"timeouts"`, using `github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts`), `GetTimeouts()` value-receiver method, and `WithResourceTimeouts` interface
- [x] 1.2 Add `ResourceTimeouts` struct with `Create, Read, Update, Delete time.Duration` fields and package-level constants `DefaultResourceCreateTimeout = 20*time.Minute`, `DefaultResourceReadTimeout = 5*time.Minute`, `DefaultResourceUpdateTimeout = 20*time.Minute`, `DefaultResourceDeleteTimeout = 20*time.Minute`
- [x] 1.3 Document the zero-value fallback semantics in godoc on `ResourceTimeouts`: each field that is zero falls back to the matching `DefaultResource<Op>Timeout` constant. The envelope reads `opts.Timeouts.<Op>` directly at call sites — no helper function or accessor methods are added (kept inline for grep-ability and to match the action-envelope precedent)
- [x] 1.4 Define a package-level `attrTimeouts = "timeouts"` constant alongside `blockElasticsearchConnection` / `blockKibanaConnection`

## 2. Elasticsearch resource envelope

- [x] 2.1 Tighten `ElasticsearchResourceModel` in `internal/entitycore/resource_envelope.go` to embed `WithResourceTimeouts`
- [x] 2.2 Add `Timeouts ResourceTimeouts` field to `ElasticsearchResourceOptions[T]`
- [x] 2.3 Store the `Timeouts` value on `ElasticsearchResource[T]` in `NewElasticsearchResource`
- [x] 2.4 Update `(r *ElasticsearchResource[T]) Schema` to inject `attrs[attrTimeouts] = timeouts.AttributesAll(ctx)` into `schema.Attributes` (mirroring the existing connection-block injection into `schema.Blocks`). Order: copy factory attributes, then overwrite the `timeouts` key — silent overwrite is the documented contract
- [x] 2.5 In `Create`, wrap ctx **after model decode and before `GetElasticsearchClient` / `EnforceVersionRequirements`** so both the client probe and the version-check API call run under the timeout: `model.GetTimeouts().Create(ctx, opts.Timeouts.Create with inline fallback to DefaultResourceCreateTimeout when zero)`, append diagnostics, return early on error, defer cancel
- [x] 2.6 In `Read`, same pattern and ordering using `.Read` and `DefaultResourceReadTimeout`, deriving timeouts from the state model
- [x] 2.7 In `Update`, same pattern and ordering using `.Update` and `DefaultResourceUpdateTimeout`, deriving timeouts from the plan model
- [x] 2.8 In `Delete`, same pattern and ordering using `.Delete` and `DefaultResourceDeleteTimeout`, deriving timeouts from the state model
- [x] 2.9 Update godoc on `ElasticsearchResource[T]`, `ElasticsearchResourceModel`, and `ElasticsearchResourceOptions[T]` to document the new timeouts contract and the silent-overwrite injection rule

## 3. Kibana resource envelope

- [x] 3.1 Tighten `KibanaResourceModel` in `internal/entitycore/kibana_resource_envelope.go` to embed `WithResourceTimeouts`
- [x] 3.2 Add `Timeouts ResourceTimeouts` to `KibanaResourceOptions[T]`
- [x] 3.3 Store on `KibanaResource[T]` in `NewKibanaResource`
- [x] 3.4 Update `(r *KibanaResource[T]) Schema` with the same attribute injection pattern as task 2.4
- [x] 3.5 Wrap ctx in `Create`, `Read`, `Update`, `Delete` (mirror tasks 2.5–2.8), placing the wrap **after model decode and before `GetKibanaClient` / `validateSpaceID` / `EnforceVersionRequirements`** so all three run under the timeout
- [x] 3.6 Update godoc to mirror task 2.9

## 4. Envelope test coverage

- [x] 4.1 Update every test model in `internal/entitycore/resource_envelope_test.go` to embed `entitycore.ResourceTimeoutsField`
- [x] 4.2 Add `resource_envelope_test.go` scenarios. Use an `httptest.Server` for the Elasticsearch backend so the test fully controls latency and response timing — the server stands in for the Stack info endpoint that `EnforceVersionRequirements` queries plus any per-op API calls:
  - schema includes a `timeouts` attribute with `create`, `read`, `update`, `delete` sub-attributes
  - configured `Options.Timeouts.Create` overrides `DefaultResourceCreateTimeout` (assert deadline propagated to the callback)
  - explicit `timeouts.create` in the plan overrides `Options.Timeouts.Create` (assert deadline matches the plan value)
  - silent overwrite: factory returns a schema whose `Attributes["timeouts"]` is a sentinel attribute; envelope's `timeouts.AttributesAll` shape wins, no panic, no diagnostic
  - ctx-wrap fires for each of the four ops (assert `ctx.Deadline()` is set inside the callback)
  - null/unknown stored `timeouts`: simulated post-upgrade state has `Timeouts` decoded as null; Read/Delete proceed under the per-op default; no diagnostics
  - version-check under timeout: httptest handler for the Stack info endpoint blocks longer than the configured op timeout; envelope returns a deadline-exceeded error before the resource callback is invoked (assert the callback was never called)
  - per-op default selection: when `Options.Timeouts.Create == 0`, deadline equals `now + DefaultResourceCreateTimeout` within a small tolerance (e.g. ±1s)
- [x] 4.3 Update every test model in `internal/entitycore/kibana_resource_envelope_test.go` to embed `ResourceTimeoutsField`
- [x] 4.4 Add Kibana envelope test scenarios mirroring task 4.2, using an `httptest.Server` standing in for the Kibana backend; cover the same eight cases plus a "space-ID validation under timeout" assertion confirming `validateSpaceID` and `EnforceVersionRequirements` both observe the wrapped deadline
- [x] 4.5 Confirm `go test ./internal/entitycore/...` passes

## 5. Mechanical model embeds — 62 resources

> Each resource model file gains one line: `entitycore.ResourceTimeoutsField` embedded alongside the existing connection-field embed. Acceptance: every resource model used with `entitycore.NewElasticsearchResource` or `entitycore.NewKibanaResource` satisfies `WithResourceTimeouts`.

- [x] 5.1 Identify every resource model file via `grep -rln "NewElasticsearchResource\|NewKibanaResource" internal/` excluding tests and the entitycore package, then locate its model struct
- [x] 5.2 Add `entitycore.ResourceTimeoutsField` embed to every identified model struct; do not modify any other field, schema, or callback
- [x] 5.3 Confirm `make build` passes (the tightened model interface compile-checks coverage)
- [x] 5.4 Run `go test ./...` to confirm no model fixture or test depending on the model struct shape is broken

## 6. Migration: `internal/fleet/customintegration`

- [x] 6.1 In `models.go`, replace the bespoke timeouts field (`Timeouts timeouts.Value` tagged `tfsdk:"timeouts"`) with `entitycore.ResourceTimeoutsField` embed; remove the now-unused `terraform-plugin-framework-timeouts/resource/timeouts` import
- [x] 6.2 In `schema.go`, delete the `"timeouts": timeouts.Attributes(ctx, timeouts.Opts{...})` entry; remove the now-unused import
- [x] 6.3 In `create.go`, delete the `plan.Timeouts.Create(ctx, 20*time.Minute)` → `context.WithTimeout` → `defer cancel()` block
- [x] 6.4 In `update.go`, delete the equivalent block
- [x] 6.5 In `resource.go` (the `NewKibanaResource` call site), pass `Timeouts: entitycore.ResourceTimeouts{Create: 20*time.Minute, Update: 20*time.Minute}`
- [x] 6.6 Run `go test ./internal/fleet/customintegration/...` (unit only — acceptance tests gated on stack availability)

## 7. Migration: `internal/elasticsearch/ml/jobstate`

- [x] 7.1 In `models.go`, replace the bespoke `Timeouts` field with `entitycore.ResourceTimeoutsField` embed
- [x] 7.2 In `schema.go`, delete the `"timeouts": timeouts.Attributes(...)` entry
- [x] 7.3 In `create.go`, delete the `data.Timeouts.Create(ctx, 5*time.Minute)` ctx-wrap block
- [x] 7.4 In `update.go`, delete the equivalent block
- [x] 7.5 In the `NewElasticsearchResource` call site, pass `Timeouts: entitycore.ResourceTimeouts{Create: 5*time.Minute, Update: 5*time.Minute}`
- [x] 7.6 Run `go test ./internal/elasticsearch/ml/jobstate/...`

## 8. Migration: `internal/elasticsearch/ml/datafeed_state`

- [x] 8.1 Mirror task 7 for `datafeed_state` — same five-minute defaults, same edit shape

## 9. Migration: `internal/elasticsearch/ml/anomalydetectionjob` (BREAKING)

- [x] 9.1 In `models_tf.go`, replace `Timeouts timeouts.Value` with `entitycore.ResourceTimeoutsField` embed
- [x] 9.2 In `schema.go`, delete the `"timeouts"` entry from the `Blocks` map returned by `getSchema()`. This is the only key in that `Blocks` map today, so the entire `Blocks: map[string]schema.Block{"timeouts": timeouts.Block(ctx, timeouts.Opts{Delete: true})}` field on the returned `schema.Schema` literal MUST be removed (not just the inner map entry — envelope injection only overwrites `Attributes`, never `Blocks`, so a stale block-key `timeouts` would coexist with the envelope's attribute and produce a duplicate-key schema error). Remove the `terraform-plugin-framework-timeouts/resource/timeouts` import that becomes unused
- [x] 9.3 In `delete.go`, delete the `state.Timeouts.Delete(ctx, 20*time.Minute)` ctx-wrap block
- [x] 9.4 In the `NewElasticsearchResource` call site, pass `Timeouts: entitycore.ResourceTimeouts{Delete: 20*time.Minute}`
- [x] 9.5 Update resource documentation (`docs/resources/elasticsearch_ml_anomaly_detection_job.md` and template under `templates/`) with a migration note: block syntax `timeouts {}` is replaced by attribute syntax `timeouts = {}`
- [x] 9.6 Add CHANGELOG entry under `BREAKING CHANGES` with before/after HCL example for `elasticsearch_ml_anomaly_detection_job`
- [x] 9.7 Run `go test ./internal/elasticsearch/ml/anomalydetectionjob/...`

## 10. Documentation regeneration and final validation

- [x] 10.1 Run `make docs-generate` (or the equivalent target) to regenerate provider docs; verify the `timeouts` attribute appears on all 66 entitycore-envelope resource pages (37 Elasticsearch + 29 Kibana)
- [x] 10.2 Run `make build`
- [x] 10.3 Run `make check-lint`
- [ ] 10.4 Run `go test ./...` (unit + envelope tests)
- [ ] 10.5 Run targeted acceptance tests for the 4 migrated resources to confirm behavior is unchanged (stack-dependent; see `dev-docs/high-level/testing.md`)
- [x] 10.6 Run `make check-openspec`
