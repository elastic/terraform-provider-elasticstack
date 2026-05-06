## 1. ML Job State — model, schema, and resource migration

- [ ] 1.1 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `MLJobStateData` in `internal/elasticsearch/ml/jobstate/models.go`.
- [ ] 1.2 Convert `Schema` receiver method to package-level `GetSchema(ctx) schema.Schema` factory omitting `elasticsearch_connection`.
- [ ] 1.3 Extract `readMLJobState(ctx, client, resourceID, state) (MLJobStateData, bool, diag.Diagnostics)` from the existing `Read` receiver method in `read.go`.
- [ ] 1.4 Change `mlJobStateResource` struct to embed `*entitycore.ElasticsearchResource[MLJobStateData]`.
- [ ] 1.5 Construct with `NewElasticsearchResource` using the schema factory, read callback, a no-op delete callback, and placeholder write callbacks.
- [ ] 1.6 Keep `Create`, `Update`, `Delete`, and `ImportState` receiver methods unchanged on the concrete type; the no-op delete callback is required for envelope construction but is shadowed by the concrete `Delete`.
- [ ] 1.7 Remove the old `Read` receiver method.

## 2. ML Datafeed State — model, schema, and resource migration

- [ ] 2.1 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `MLDatafeedStateData` in `internal/elasticsearch/ml/datafeed_state/models.go`.
- [ ] 2.2 Convert `Schema` receiver method to package-level `GetSchema(ctx) schema.Schema` factory omitting `elasticsearch_connection`.
- [ ] 2.3 Extract `readMLDatafeedState(ctx, client, resourceID, state) (MLDatafeedStateData, bool, diag.Diagnostics)` from the existing `read` helper in `read.go`.
- [ ] 2.4 Extract the body of the existing `Delete` receiver method into a package-level `deleteMLDatafeedState(ctx, client, resourceID, state) diag.Diagnostics` callback.
- [ ] 2.5 Change `mlDatafeedStateResource` struct to embed `*entitycore.ElasticsearchResource[MLDatafeedStateData]`.
- [ ] 2.6 Construct with `NewElasticsearchResource` using the schema factory, read callback, delete callback, and placeholder write callbacks.
- [ ] 2.7 Keep `Create`, `Update`, and `ImportState` receiver methods unchanged on the concrete type. Update the concrete `Delete` receiver to delegate to `deleteMLDatafeedState` so the extracted helper is the single implementation of conditional stop behavior.
- [ ] 2.8 Remove the old `Read` receiver method (and ensure `read` helper is used by both the callback and the override).

## 3. Verification

- [ ] 3.1 Run `make build`.
- [ ] 3.2 Run `make check-lint`.
- [ ] 3.3 Run `make check-openspec`.
- [ ] 3.4 Run acceptance tests for `ml_job_state` and `ml_datafeed_state` if infrastructure is available.
