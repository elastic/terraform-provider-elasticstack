## Context

The provider is incrementally migrating Elasticsearch API calls from the raw `esapi` (untyped) client to the `go-elasticsearch` Typed API (`elasticsearch.TypedClient` via `ToTyped()`). Previous phases migrated lower-level helpers; this change targets the cluster surface in `internal/clients/elasticsearch/cluster.go`.

Current state:
- `GetClusterInfo` uses `esClient.Info()` and decodes into `models.ClusterInfo`.
- Snapshot repository, SLM, settings, and script helpers manually marshal/unmarshal JSON via `esapi` and use custom models (`models.SnapshotRepository`, `models.SnapshotPolicy`, `models.Script`).
- `internal/clients/elasticsearch/helpers.go` provides `doSDKWrite`/`doFWWrite` wrappers for raw API calls.
- Callers in `internal/elasticsearch/cluster/` (resource and test files) consume the helper return types directly.
- `ElasticsearchScopedClient` caches `*models.ClusterInfo` in `serverInfo()` and exposes `ClusterID()`, `ServerVersion()`, `ServerFlavor()`.

## Goals / Non-Goals

**Goals:**
- Migrate all helpers in `internal/clients/elasticsearch/cluster.go` to typed API calls.
- Remove custom models that duplicate upstream typed API structs (`ClusterInfo`, `SnapshotRepository`, `SnapshotPolicy`, `Script`).
- Update all callers to use typed API response types.
- Add a typed client accessor on `ElasticsearchScopedClient` so helpers can consistently obtain `*elasticsearch.TypedClient`.
- Ensure the `serverInfo()` cache continues to work with typed responses.
- Keep zero Terraform-level behavioral changes.

**Non-Goals:**
- Migrating other files in `internal/clients/elasticsearch/`.
- Changing resource schemas, documentation, or Terraform user-facing behavior.
- Adding new resources or data sources.

## Decisions

### 1. Typed client access via `ElasticsearchScopedClient.GetESTypedClient()`

**Decision:** Add `GetESTypedClient() (*elasticsearch.TypedClient, error)` to `ElasticsearchScopedClient`.

**Rationale:** Helpers currently call `GetESClient()` and then use the raw `*elasticsearch.Client`. To use the typed API they need `*elasticsearch.TypedClient`, obtained via `client.ToTyped()`. Centralizing this in the scoped client ensures endpoint validation (already present in `GetESClient()`) and avoids repeated `ToTyped()` calls scattered through helpers.

**Alternative considered:** Call `ToTyped()` inline inside each helper. Rejected because it duplicates endpoint-not-configured checks and makes it harder to inject typed client mocks later.

### 2. Return typed API response types instead of custom models

**Decision:** Change return types:
- `GetClusterInfo` → `*core.InfoResponse`
- `GetSnapshotRepository` → `getrepository.Response` (map of `types.Repository`)
- `GetSlm` → `*types.SLMPolicy` (extracted from `getlifecycle.Response`)
- `GetScript` → `*types.StoredScript`
- `GetSettings` → `*getsettings.Response`

**Rationale:** Eliminates custom models that duplicate upstream structs and require manual JSON decoding. The typed API handles marshaling internally.

**Impact on callers:** Callers need minor adaptation:
- `cluster_info_data_source.go` reads `info.Version.Int` instead of `info.Version.Number` (the typed API uses `Int` for the version number field).
- `snapshot_repository.go` and `snapshot_repository_data_source.go` iterate over `types.Repository` union values; each branch must be type-asserted to the concrete repository type (e.g., `types.SharedFileSystemRepository`, `types.S3Repository`, etc.) to access `Type` and `Settings`.
- `slm.go` directly uses `types.SLMPolicy` and `types.Retention`.
- `script` resource uses `types.StoredScript` (fields: `Lang`, `Source` — note `Params` is not part of `StoredScript`; callers must handle params separately).
- `settings.go` handles `json.RawMessage` values in `getsettings.Response`.

### 3. Keep helper signatures in `cluster.go` but change internals

**Decision:** Retain the same function names and parameter lists; only migrate bodies and return types.

**Rationale:** Minimizes diff size and avoids renaming churn across dozen+ call sites in resources and tests.

### 4. Remove obsolete custom models from `models.go`

**Decision:** Delete `ClusterInfo`, `SnapshotRepository`, `SnapshotPolicy`/`SnapshortRetention`/`SnapshotPolicyConfig`, and `Script` from `internal/models/models.go` after migrating all consumers.

**Rationale:** Prevents drift between custom and upstream types. These models exist solely to bridge the raw API; the typed API makes them unnecessary.

### 5. Adapt `ElasticsearchScopedClient.serverInfo()` to typed API

**Decision:** Rewrite `serverInfo()` to use `typedClient.Core.Info().Do(ctx)`, returning `*core.InfoResponse`, and update the cached field type.

**Rationale:** The cache uses `*models.ClusterInfo` today; it must switch to `*core.InfoResponse` so downstream methods (`ClusterID`, `ServerVersion`, `ServerFlavor`) compile without conversion layers.

**Trade-off:** `core.InfoResponse` stores `BuildDate` as `types.DateTime` (not a custom `time.Time` wrapper). `DateTime` has its own `String()` implementation, so the data source `build_date` output may change format if the upstream stringer differs from the old custom `UnmarshalJSON` + `time.Time.String()`. Both produce RFC3339-like output; this is acceptable for a pure refactor.

### 6. Snapshot repository typed API union handling

**Decision:** In `GetSnapshotRepository`, after obtaining `getrepository.Response`, type-switch on `types.Repository` to extract the concrete repository struct and read its `Type` and `Settings` fields.

**Rationale:** `types.Repository` is a Go `any` alias covering `AzureRepository`, `GcsRepository`, `S3Repository`, `SharedFileSystemRepository`, `ReadOnlyUrlRepository`, and `SourceOnlyRepository`. The existing resource only needs `Type` (discriminator) and the raw `Settings` map. Each concrete type exposes `Type` (string) and `Settings` (map[string]json.RawMessage or similar); a small adapter can flatten the settings into the `map[string]any` shape the resource already expects.

## Risks / Trade-offs

- **[Risk]** `types.Repository` union type-assertion logic is new code; missed branches could panic at runtime.
  → **Mitigation:** Use a type switch with an explicit default returning a diagnostic error. Cover all known repository types (`fs`, `url`, `gcs`, `azure`, `s3`, `hdfs`, `source`); add a catch-all error for unknown types.

- **[Risk]** `getsettings.Response` fields are `map[string]json.RawMessage`, requiring `json.Unmarshal` per value where the old code decoded into `map[string]any` in one step.
  → **Mitigation:** Add a small flattening helper in `cluster.go` that unmarshals each `RawMessage` into `any` and returns the aggregated `map[string]any`. This preserves the existing caller contract.

- **[Risk]** Script `params` field is not present on `types.StoredScript`; only `Lang`, `Source`, and `Options` exist.
  → **Mitigation:** Keep the script resource's internal `params` logic (marshal/unmarshal from Terraform state) unchanged. `PutScript` will continue building the request body that includes params via a wrapper struct, but instead of using `models.Script`, it uses `types.StoredScript` plus params. Alternatively, build the request via the typed API `PutScript` request builder which accepts `*types.StoredScript` directly and handle params as an additional state-only field.

- **[Risk]** Build date formatting change in `cluster_info_data_source.go`.
  → **Mitigation:** Verify in acceptance tests that `build_date` remains a valid timestamp string; adapt if the exact string format changes.

## Migration Plan

1. Add `GetESTypedClient()` to `ElasticsearchScopedClient`.
2. Update `serverInfo()` and cache to use `*core.InfoResponse`.
3. Rewrite cluster helpers (`GetClusterInfo`, Snapshot repository, SLM, Settings, Script) to typed API.
4. Update callers:
   - `cluster_info_data_source.go`
   - `snapshot_repository.go` + `snapshot_repository_data_source.go`
   - `slm.go`
   - `settings.go`
   - `script/` resource files (`create.go`, `read.go`, `delete.go`, `update.go`)
5. Remove dead models from `models.go`.
6. Run `make build` and targeted acceptance tests.
7. Archive change or sync delta specs if necessity is discovered.

## Open Questions

(none)
