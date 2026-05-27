## 1. Provider registration

- [x] 1.1 In `provider/plugin_framework.go`, add `res.ActionData = factory` in `Configure()` alongside the existing `res.ResourceData`, `res.DataSourceData`, and `res.EphemeralResourceData` assignments.
- [x] 1.2 Add `_ fwprovider.ProviderWithActions = &Provider{}` to the `var` block of interface assertions.
- [x] 1.3 Add `Actions(ctx context.Context) []func() action.Action` method on `*Provider` that returns both action constructors (to be added in later tasks).
- [x] 1.4 Add the required import `"github.com/hashicorp/terraform-plugin-framework/action"` and `fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"` (provider import alias already present).

## 2. Snapshot restore — client helper

- [x] 2.1 Create `internal/clients/elasticsearch/snapshot_restore.go` with a `RestoreSnapshot` function. Signature: `func RestoreSnapshot(ctx context.Context, client *clients.ElasticsearchScopedClient, repo, snapshot string, body *RestoreSnapshotRequest, waitForCompletion bool) fwdiag.Diagnostics`.
- [x] 2.2 Define `RestoreSnapshotRequest` struct with fields matching `POST /_snapshot/{repo}/{snapshot}/_restore` body: `Indices []string`, `IgnoreUnavailable *bool`, `IncludeGlobalState *bool`, `IncludeAliases *bool`, `FeatureStates []string`, `RenamePattern *string`, `RenameReplacement *string`, `Partial *bool`, `IndexSettings json.RawMessage` (JSON object), `IgnoreIndexSettings []string`.
- [x] 2.3 Decode `IndexSettings` from `jsontypes.Normalized` into `json.RawMessage`/`map[string]any` (or equivalent typed API object) before building the request body so ES receives an object, not an escaped string.
- [x] 2.4 Implement `RestoreSnapshot` using the typed ES client's `typedapi/snapshot/restore` package. Pass `wait_for_completion` as the query parameter. Propagate ES errors as framework diagnostics.

## 3. Snapshot create — client helper

- [x] 3.1 Create `internal/clients/elasticsearch/snapshot_create.go` with a `CreateSnapshot` function. Signature: `func CreateSnapshot(ctx context.Context, client *clients.ElasticsearchScopedClient, repo, snapshot string, body *CreateSnapshotRequest, waitForCompletion bool) fwdiag.Diagnostics`.
- [x] 3.2 Define `CreateSnapshotRequest` struct with fields: `Indices []string`, `IgnoreUnavailable *bool`, `IncludeGlobalState *bool`, `FeatureStates []string`, `ExpandWildcards *string`, `Metadata json.RawMessage` (JSON object), `Partial *bool`.
- [x] 3.3 Decode `Metadata` from `jsontypes.Normalized` into `json.RawMessage`/`map[string]any` (matching the SLM metadata mapping approach) before building the request body so ES receives an object, not an escaped string.
- [x] 3.4 Implement `CreateSnapshot` using `typedapi/snapshot/create`. Pass `wait_for_completion` as the query parameter. Propagate ES errors as framework diagnostics.

## 4. Snapshot restore — action package

- [x] 4.1 Create package `internal/elasticsearch/cluster/snapshot_restore/`.
- [x] 4.2 Create `model.go` with `Model` struct (framework types): `Repository types.String`, `Snapshot types.String`, `Indices types.List`, `IncludeGlobalState types.Bool`, `IgnoreUnavailable types.Bool`, `IncludeAliases types.Bool`, `Partial types.Bool`, `FeatureStates types.List`, `RenamePattern types.String`, `RenameReplacement types.String`, `IgnoreIndexSettings types.List`, `IndexSettings jsontypes.Normalized`, `WaitForCompletion types.Bool`, `Timeouts timeouts.Value` (from `action/timeouts`), `ElasticsearchConnection types.List`.
- [x] 4.3 Create `schema.go` with `GetSchema(ctx context.Context) action.Schema`. Include all model fields, plus the `elasticsearch_connection` block (reuse existing schema helper). Add `timeouts` block using `action/timeouts` package. Mark `repository` and `snapshot` as required.
- [x] 4.4 Create `action.go` with `snapshotRestoreAction` struct implementing `action.Action` and `action.ActionWithConfigure`:
  - `Metadata`: sets `TypeName = "elasticstack_elasticsearch_snapshot_restore"`
  - `Schema`: delegates to `GetSchema`
  - `Configure`: extracts `*clients.ProviderClientFactory` from `req.ProviderData` using `clients.ConvertProviderDataToFactory`
  - `Invoke`: reads `Model` from `req.Config`; resolves scoped ES client; applies `invoke` timeout from `timeouts` block via `context.WithTimeout`; calls `elasticsearch.RestoreSnapshot`; appends diagnostics
- [x] 4.5 Create `NewRestoreAction() func() action.Action` constructor function in `action.go`.

## 5. Snapshot create — action package

- [x] 5.1 Create package `internal/elasticsearch/cluster/snapshot_create/`.
- [x] 5.2 Create `model.go` with `Model` struct: `Repository types.String`, `Snapshot types.String`, `Indices types.List`, `IncludeGlobalState types.Bool`, `IgnoreUnavailable types.Bool`, `Partial types.Bool`, `FeatureStates types.List`, `ExpandWildcards types.String`, `Metadata jsontypes.Normalized`, `WaitForCompletion types.Bool`, `Timeouts timeouts.Value`, `ElasticsearchConnection types.List`.
- [x] 5.3 Create `schema.go` with `GetSchema(ctx context.Context) action.Schema`. Include all model fields plus `elasticsearch_connection` block and `timeouts` block (using `action/timeouts`).
- [x] 5.4 Create `action.go` with `snapshotCreateAction` struct implementing `action.Action` and `action.ActionWithConfigure`:
  - `Metadata`: `TypeName = "elasticstack_elasticsearch_snapshot_create"`
  - `Schema`, `Configure`, `Invoke`: same pattern as snapshot_restore action
- [x] 5.5 Create `NewCreateAction() func() action.Action` constructor in `action.go`.

## 6. Register actions in provider

- [x] 6.1 Import `snapshot_restore` and `snapshot_create` packages in `provider/plugin_framework.go`.
- [x] 6.2 Return `[]func() action.Action{snapshot_restore.NewRestoreAction, snapshot_create.NewCreateAction}` from the `Actions()` method.

## 7. Build verification

- [x] 7.1 Run `make build` and confirm no compilation errors.

## 8. Acceptance tests — snapshot restore

- [x] 8.1 Create `internal/elasticsearch/cluster/snapshot_restore/acc_test.go` with `TestAccActionSnapshotRestore`.
- [x] 8.2 Test config (in `testdata/TestAccActionSnapshotRestore/`): create a filesystem snapshot repository, create an index with test data, create a snapshot via `elasticstack_elasticsearch_snapshot_create` action, then invoke the restore action with `rename_pattern`/`rename_replacement` to avoid conflicts.
- [x] 8.3 Assert the restore completes without errors (action has no output attributes; check that diagnostics are clean) and verify via Elasticsearch that the renamed restored index exists.
- [x] 8.4 Test `wait_for_completion = false` path to confirm action proceeds without waiting.

## 9. Acceptance tests — snapshot create

- [x] 9.1 Create `internal/elasticsearch/cluster/snapshot_create/acc_test.go` with `TestAccActionSnapshotCreate`.
- [x] 9.2 Test config: create a filesystem snapshot repository, then invoke the create action for a set of indices.
- [x] 9.3 Assert creation completes without errors and verify via the snapshot API that the snapshot exists; include a metadata case and assert metadata is attached.
- [x] 9.4 Test `wait_for_completion = false` path.

## 10. Documentation

- [x] 10.1 Add `Requires Terraform 1.14+` note to the description/template for both action docs pages.
- [x] 10.2 Run `make docs` (or equivalent) to regenerate provider docs if needed.
