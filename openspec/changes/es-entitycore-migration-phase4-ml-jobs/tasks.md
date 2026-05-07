## 1. Anomaly Detection Job — model and schema

- [ ] 1.1 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `TFModel` in `internal/elasticsearch/ml/anomalydetectionjob/models.go` (or wherever the struct is defined).
- [ ] 1.2 Convert `func (r *anomalyDetectionJobResource) Schema` to a package-level factory `func getSchema() schema.Schema` omitting the `elasticsearch_connection` block.

## 2. Anomaly Detection Job — callback extraction

- [ ] 2.1 Extract `readAnomalyDetectionJob(ctx, client, resourceID, state) (TFModel, bool, diag.Diagnostics)` from the existing `read` receiver method.
- [ ] 2.2 Extract `createAnomalyDetectionJob(ctx, client, resourceID, plan) (TFModel, diag.Diagnostics)` from the existing `create` receiver method (remove the read-after-write and state.Set; just create, set ID, and return).
- [ ] 2.3 Extract `deleteAnomalyDetectionJob(ctx, client, resourceID, state) diag.Diagnostics` from the existing `delete` receiver method.
- [ ] 2.4 Keep the `update` receiver method logic in place; it will become the body of the overridden `Update` method.

## 3. Anomaly Detection Job — resource struct migration

- [ ] 3.1 Change `anomalyDetectionJobResource` struct to embed `*entitycore.ElasticsearchResource[TFModel]`.
- [ ] 3.2 Construct it with `NewElasticsearchResource` using the schema factory, `readAnomalyDetectionJob`, `deleteAnomalyDetectionJob`, `createAnomalyDetectionJob`, and a placeholder update callback.
- [ ] 3.3 Implement `Update` on the concrete type: decode plan and state, build update body, send raw JSON update, read back, set state.
- [ ] 3.4 Keep `ImportState` on the concrete type unchanged.
- [ ] 3.5 Remove the old `Read`, `Create`, and `Delete` receiver methods.

## 4. Datafeed — model and schema

- [ ] 4.1 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `Datafeed` in `internal/elasticsearch/ml/datafeed/models.go`.
- [ ] 4.2 Convert `func (r *datafeedResource) Schema` to a package-level factory `func getSchema() schema.Schema` omitting the `elasticsearch_connection` block.

## 5. Datafeed — callback extraction

- [ ] 5.1 Extract `readDatafeed(ctx, client, resourceID, state) (Datafeed, bool, diag.Diagnostics)`.
- [ ] 5.2 Extract `createDatafeed(ctx, client, resourceID, plan) (Datafeed, diag.Diagnostics)` (remove read-after-write and state.Set; just create, set ID, and return).
- [ ] 5.3 Extract `updateDatafeed(ctx, client, resourceID, plan) (Datafeed, diag.Diagnostics)` (remove read-after-write and state.Set; stop/update/start, set ID, and return).
- [ ] 5.4 Extract `deleteDatafeed(ctx, client, resourceID, state) diag.Diagnostics`.

## 6. Datafeed — resource struct migration

- [ ] 6.1 Change `datafeedResource` struct to embed `*entitycore.ElasticsearchResource[Datafeed]`.
- [ ] 6.2 Construct it with `NewElasticsearchResource` using all four real callbacks.
- [ ] 6.3 Keep `ImportState` on the concrete type unchanged.
- [ ] 6.4 Remove the old `Create`, `Read`, `Update`, and `Delete` receiver methods.

## 7. Verification

- [ ] 7.1 Run `make build`.
- [ ] 7.2 Run `make check-lint`.
- [ ] 7.3 Run `make check-openspec`.
- [ ] 7.4 Run focused unit tests for both packages.
- [ ] 7.5 Run acceptance tests for `ml_anomaly_detection_job` and `ml_datafeed` if infrastructure is available.
