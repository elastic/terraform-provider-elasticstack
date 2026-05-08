## 1. Plugin Framework model and schema

- [x] 1.1 Define `tfModel` struct in `internal/elasticsearch/transform/` with all PF types matching the existing SDK schema. Include nested block model types for `source`, `destination`, `retention_policy`, `sync`, and `aliases`.
- [x] 1.2 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `tfModel`.
- [x] 1.3 Write `getSchema() schema.Schema` factory omitting `elasticsearch_connection`. Include:
  - `stringvalidator.ExactlyOneOf` for `pivot`/`latest`
  - `RequiresReplace` plan modifiers on `name`, `pivot`, `latest`
  - `jsontypes.Normalized` for JSON fields
  - Custom validators for name, index, duration fields
  - `UseStateForUnknown` on computed attributes

## 2. Model conversion helpers

- [x] 2.1 Write `toAPIModel(ctx, model, serverVersion) (*models.Transform, diag.Diagnostics)` that converts the PF model to the API request struct, applying version gating via `isSettingAllowed` checks.
- [x] 2.2 Write `fromAPIModel(ctx, transform, stats) (tfModel, diag.Diagnostics)` that populates the PF model from the Get Transform response and Get Transform Stats.

## 3. Read and delete callbacks

- [x] 3.1 Implement `readTransform(ctx, client, resourceID, state) (tfModel, bool, diag.Diagnostics)`:
  - Call `elasticsearch.GetTransform`
  - Return `found=false` on 404
  - Call `elasticsearch.GetTransformStats`
  - Derive `enabled` from stats state
  - Populate model and return
- [x] 3.2 Implement `deleteTransform(ctx, client, resourceID, state) diag.Diagnostics` that calls `elasticsearch.DeleteTransform` with `force=true`.

## 4. Create callback

- [x] 4.1 Implement `createTransform(ctx, client, resourceID, model) (tfModel, diag.Diagnostics)`:
  - Resolve server version
  - Convert model to API request via `toAPIModel`
  - Call `elasticsearch.PutTransform` with `defer_validation` and `timeout`
  - If `model.Enabled`, call `elasticsearch.StartTransform` with `timeout`
  - Build composite `id` and set it on the model
  - Return the model

## 5. Resource struct and update override

- [x] 5.1 Define `type transformResource struct { *entitycore.ElasticsearchResource[tfModel] }`.
- [x] 5.2 Implement exported `NewTransformResource() resource.Resource` constructing the envelope with schema factory, read callback, delete callback, create callback, and placeholder update callback.
- [x] 5.3 Implement `Update` override:
  - Decode plan and state into separate models
  - Build update request from plan, omitting `pivot` and `latest`
  - Call `elasticsearch.UpdateTransform` with `defer_validation` and `timeout`
  - If `enabled` changed from `false` to `true`, call `elasticsearch.StartTransform`
  - If `enabled` changed from `true` to `false`, call `elasticsearch.StopTransform`
  - Call `readTransform` callback to refresh state
  - Persist refreshed model to `resp.State`
- [x] 5.4 Implement `ImportState` as passthrough on `id`.

## 6. Provider wiring and cleanup

- [x] 6.1 Replace the SDK `ResourceTransform()` registration in the provider with the new PF `NewTransformResource()` factory.
- [x] 6.2 Remove the old SDK resource code once the PF version compiles.
- [x] 6.3 Update any provider-level type assertions or resource lists.

## 7. Verification

- [x] 7.1 Run `make build`.
- [x] 7.2 Run `make check-lint`.
- [x] 7.3 Run `make check-openspec`.
- [x] 7.4 Run focused tests for `internal/elasticsearch/transform`.
- [x] 7.5 Run acceptance tests for `transform` if infrastructure is available.
