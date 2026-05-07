## 1. Migrate `elasticstack_elasticsearch_watch` to envelope

- [ ] 1.1 Add value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` to `internal/elasticsearch/watcher/watch/models.go` `Data`.
- [ ] 1.2 Refactor the existing `read` method in `internal/elasticsearch/watcher/watch/read.go` into a package-level `readWatch(ctx, client, resourceID string, state Data) (Data, bool, diag.Diagnostics)` callback. The body SHALL: parse composite ID from `state.ID`, get client, call `elasticsearch.GetWatch`, invoke `state.fromAPIModel(ctx, watch, state.Actions)`, copy `state.ID` and `state.ElasticsearchConnection` into the result, and return. Return `(_, false, nil)` when watch is nil.
- [ ] 1.3 Remove the `Read` method from `internal/elasticsearch/watcher/watch/resource.go`.
- [ ] 1.4 Extract `createWatch` callback from the existing `Create` method body in `internal/elasticsearch/watcher/watch/create.go`. Signature: `(ctx, client, resourceID string, plan Data) (Data, diag.Diagnostics)`. Body SHALL: build `put` model via `plan.toPutModel`, call `elasticsearch.PutWatch`, compute composite id via `client.ID(ctx, plan.WatchID.ValueString())`, set `plan.ID`, and return `plan`. Do NOT call `read` inside the callback.
- [ ] 1.5 Extract `updateWatch` callback from the existing `Update` method body. Same signature as create. Body is the same PUT flow, except it SHALL build the request body with update semantics: when `transform` is not configured, include `transform: {}` so Elasticsearch clears any existing transform.
- [ ] 1.6 Remove the `Create` and `Update` methods from `resource.go`.
- [ ] 1.7 Extract `deleteWatch` callback from the existing `Delete` method body in `internal/elasticsearch/watcher/watch/delete.go`. Signature: `(ctx, client, resourceID string, state Data) diag.Diagnostics`. Body calls `elasticsearch.DeleteWatch`.
- [ ] 1.8 Remove the `Delete` method from `resource.go`.
- [ ] 1.9 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `watchResource`. In `newWatchResource()`, call `entitycore.NewElasticsearchResource[Data]` with component `"watch"`, schema factory, `readWatch`, `deleteWatch`, `createWatch`, and `updateWatch`.
- [ ] 1.10 Strip the `elasticsearch_connection` block from `schema.go`; the envelope injects it.
- [ ] 1.11 Preserve `ImportState` passthrough on the concrete `watchResource` type.
- [ ] 1.12 Update interface assertions in `resource.go` as needed.

## 2. Verification

- [ ] 2.1 `make build` passes.
- [ ] 2.2 `make check-lint` passes.
- [ ] 2.3 `make check-openspec` passes.
- [ ] 2.4 Unit tests in `internal/elasticsearch/watcher/watch/` pass.
- [ ] 2.5 Acceptance tests for `elasticstack_elasticsearch_watch` pass against a running stack.
- [ ] 2.6 Confirm redacted actions acceptance tests still pass (redaction round-trip with `::es_redacted::`).
