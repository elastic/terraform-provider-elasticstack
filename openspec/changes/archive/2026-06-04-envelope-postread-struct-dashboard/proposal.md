## Why

Both `KibanaPostReadFunc[T]` and `PostReadFunc[T]` (the Elasticsearch envelope equivalent) currently receive only the freshly-read model and a `privateState any` value. Neither has access to the model that existed before the read — making it impossible for post-read logic that needs to compare the server's response against the prior desired or persisted state (e.g. panel layout alignment in `kibana_dashboard`). Changing both callback signatures to struct-based requests (matching the pattern of `KibanaWriteFunc[T]`) provides `Prior T` (the plan on write paths, the prior state on read paths) alongside `State T` and `Private`, unblocking the `kibana/dashboard` migration and giving both envelopes a uniform, extensible PostRead contract.

## What Changes

- **`internal/entitycore/kibana_resource_envelope.go`**: Replace the current `KibanaPostReadFunc[T]` with a struct-based callback:
  - Introduce `KibanaPostReadRequest[T]` with fields `Client *clients.KibanaScopedClient`, `Prior T`, `State T`, `Private any`
  - `Prior` carries the write-request plan on Create/Update paths; on plain Read it carries the state model before this refresh
  - Change `KibanaPostReadFunc[T]` to `func(ctx context.Context, req KibanaPostReadRequest[T]) (T, diag.Diagnostics)` — returns the (possibly modified) model; envelope sets state after PostRead returns
  - Update `runKibanaRead` and `runKibanaWrite` to populate `Prior`, pass the struct, and apply the returned model before setting state
- **`internal/entitycore/resource_envelope.go`**: Apply the same treatment to `PostReadFunc[T]`:
  - Introduce `ElasticsearchPostReadRequest[T]` with the same four fields, `Client *clients.ElasticsearchScopedClient`
  - Change `PostReadFunc[T]` to `func(ctx context.Context, req ElasticsearchPostReadRequest[T]) (T, diag.Diagnostics)`
  - Update `runElasticsearchRead` and `runElasticsearchWrite` accordingly
- **`internal/elasticsearch/security/apikey/resource/resource.go`**: Update `postReadPersistAPIKeyCapabilities` to the new `PostReadFunc[T]` signature; function ignores the model, so it returns `(req.State, diags)` unchanged
- **`internal/entitycore/kibana_resource_envelope_test.go`** and **`internal/entitycore/resource_envelope_test.go`**: Update all existing PostRead test lambdas to the new signatures; add new scenarios covering `Prior` population and state-set-from-PostRead-return behaviour
- **`internal/kibana/dashboard`**: Migrate from `*entitycore.ResourceBase` to `*entitycore.KibanaResource[DashboardModel]`, using the new `KibanaPostReadFunc[T]` with `req.Prior` to implement all four panel/section/pinned-panel alignment functions (currently scattered across `create.go`, `update.go`, and `read.go`) in a single `postReadDashboard` callback
  - Add `KibanaResourceModel` interface methods to `DashboardModel`
  - Remove the now-redundant explicit `kibana_connection` block from `schema.go` (envelope injects it)
  - Rewrite create, update, read, and delete as envelope callbacks
  - `read.go` callback: raw API call and model population only — no alignment
  - `PostRead`: all four alignment functions using `req.Prior` as the reference; works correctly for both write path (Prior = plan) and plain Read path (Prior = prior state)
  - Update `entitycore_contract_test.go` (already exists; asserts `ResourceBase` — must be updated to assert `KibanaResource[DashboardModel]`)

## Capabilities

### New Capabilities

None — `kibana_dashboard` already exists; no schema or behavioural changes visible to practitioners.

### Modified Capabilities

None — the `entitycore-kibana-resource-envelope` and `entitycore-resource-envelope` internal contracts change but no externally specified capability requirements change.

## Impact

- `internal/entitycore/kibana_resource_envelope.go` — new struct, changed type, updated `runKibanaRead` and `runKibanaWrite`
- `internal/entitycore/kibana_resource_envelope_test.go` — 7 PostRead test functions updated; new test scenarios added
- `internal/entitycore/resource_envelope.go` — new struct, changed type, updated `runElasticsearchRead` and `runElasticsearchWrite`
- `internal/entitycore/resource_envelope_test.go` — 7 PostRead test functions updated; new test scenarios added
- `internal/elasticsearch/security/apikey/resource/resource.go` — one function signature update
- `internal/kibana/dashboard/` — all non-test Go files; `DashboardModel` gains interface methods; `entitycore_contract_test.go` updated
- No schema changes; acceptance tests unaffected
