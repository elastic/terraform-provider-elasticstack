## Context

Both `anomalydetectionjob` and `datafeed` are Plugin Framework resources with well-factored CRUD implementations. Each has:
- A `resource.go` file with thin `Create`/`Read`/`Update`/`Delete` wrappers and custom `ImportState`.
- Separate `create.go`, `update.go`, `delete.go`, and `read.go` files containing the actual Elasticsearch API logic.
- Typed-client API usage (`Ml.PutJob`, `Ml.UpdateJob`, `Ml.PutDatafeed`, `Ml.UpdateDatafeed`, etc.).

The anomaly detection job's update logic is unique: it compares plan and state to build a partial update body containing only changed, mutable fields. The datafeed's update logic does not need the prior Terraform state (it checks live API state to decide stop/restart), so it fits the envelope callback contract.

## Goals / Non-Goals

**Goals:**

- Migrate both resources to the entitycore envelope.
- Preserve all existing API behavior, schema, and acceptance tests.
- Keep custom ImportState on both concrete types.

**Non-Goals:**

- Changing the anomaly detection job update logic.
- Changing the datafeed stop/restart behavior.
- Changing schemas or acceptance tests.

## Decisions

### D1. Anomaly detection job overrides Update; datafeed uses full callbacks

**Choice:**
- `anomaly_detection_job`: real create callback, real read callback, real delete callback, overridden Update.
- `datafeed`: real callbacks for create, read, update, and delete.

**Rationale:** The anomaly detection job's `BuildFromPlan` compares the plan and state models to determine which fields changed. The envelope's update callback only receives the planned model, so it cannot perform this comparison. The datafeed update does not reference prior Terraform state at all; it only checks live API state, so it fits the callback contract cleanly.

### D2. Read callbacks return bool for found

**Choice:** Both read callbacks return `(T, bool, diag.Diagnostics)` where bool signals found/not-found. The envelope handles `RemoveResource` when `false`.

**Rationale:** Both current `read` helpers already return a found boolean. Aligning with the envelope callback signature means the concrete `Read` wrappers can be deleted entirely.

### D3. ImportState stays on concrete types

**Choice:** Both resources implement `ResourceWithImportState` themselves.

**Rationale:** The envelope does not implement ImportState (opt-in convention). The anomaly detection job uses a composite-ID parse that sets both `id` and `job_id`. The datafeed uses passthrough plus composite-ID parse that sets `datafeed_id`.

### D4. Schema factories strip connection block

**Choice:** Both schema factories return `schema.Schema` without `elasticsearch_connection`. The envelope injects it.

**Rationale:** Standard envelope convention.

### D5. GetResourceID returns the natural write identity

**Choice:** `TFModel.GetResourceID()` returns `JobID`. `Datafeed.GetResourceID()` returns `DatafeedID`.

**Rationale:** These are the plan-safe natural identifiers. The computed `id` is built from them after a successful API call.

## Risks / Trade-offs

- **Risk:** The anomaly detection job update override duplicates a small amount of envelope prelude (plan decode, client resolution). **Mitigation:** The override is ~8 lines; the actual update logic stays in a shared helper.
- **Risk:** Datafeed create/update currently set `plan.ID` and then call `read`, then set state. Under the envelope, the callback sets `ID` on the returned model, the envelope calls `readFunc`, and the envelope sets state. **Mitigation:** The datafeed callback must set `ID` before returning; the read callback will refresh the full model.

## Migration Plan

1. For `anomalydetectionjob`:
   - Add getters to `TFModel`.
   - Extract `readAnomalyDetectionJob`, `createAnomalyDetectionJob`, `deleteAnomalyDetectionJob` as package-level callbacks.
   - Convert schema to a factory.
   - Replace `*ResourceBase` with `*entitycore.ElasticsearchResource[TFModel]`.
   - Keep `Create` and `Delete` as envelope callbacks.
   - Override `Update` on the concrete type.
   - Keep `ImportState` on the concrete type.

2. For `datafeed`:
   - Add getters to `Datafeed`.
   - Extract `readDatafeed`, `createDatafeed`, `updateDatafeed`, `deleteDatafeed` as package-level callbacks.
   - Convert schema to a factory.
   - Replace `*ResourceBase` with `*entitycore.ElasticsearchResource[Datafeed]`.
   - Use full envelope CRUD.
   - Keep `ImportState` on the concrete type.

3. Run `make build`, `make check-lint`, `make check-openspec`, and acceptance tests.

## Open Questions

None.
