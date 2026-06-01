## Context

The Elasticsearch ML trained model alias API (`PUT`/`DELETE _ml/trained_models/{model_id}/model_aliases/{model_alias}`) enables logical names for trained models. Aliases decouple inference pipelines from specific model IDs, allowing seamless model upgrades by reassigning the alias rather than updating every consumer. This resource follows the same `entitycore.NewElasticsearchResource[T]` envelope pattern used by other ML resources (`filter`, `datafeed`, `anomalydetectionjob`).

API reference:
- Create/Update: `PUT /_ml/trained_models/{model_id}/model_aliases/{model_alias}?reassign={bool}`
- Read: `GET /_ml/trained_models/{model_alias}` (GetTrainedModels with alias as the model_id parameter)
- Delete: `DELETE /_ml/trained_models/{model_id}/model_aliases/{model_alias}`
- Required cluster privilege: `manage_ml`

## Goals / Non-Goals

**Goals:**
- Full CRUD resource for ML trained model aliases using the `entitycore` envelope.
- `model_alias` is the stable resource identity (ForceNew, composite-id key).
- `model_id` is the current referent; it is **not** ForceNew — it can be updated in-place via PUT with `reassign=true`, as described in the drift-handling section of the issue.
- `reassign` controls whether the PUT succeeds when the alias already exists pointing to a different model.
- Drift: alias deleted out-of-band → read returns empty → re-create on next apply. Alias reassigned out-of-band → model_id mismatch shows as a diff → update on next apply.
- Import by composite `<cluster_uuid>/<model_alias>`.
- Acceptance tests covering create, update (reassign), no-diff, import, and delete.

**Non-Goals:**
- Listing or enumerating aliases (no data source planned in this change).
- Managing the model itself (separate resource `elasticstack_elasticsearch_ml_trained_model`).
- Cross-type reassignment (ES API rejects reassigning an alias from a regression model to a classification model; the resource surfaces this as an API error, not a pre-flight check).

## Decisions

### 1. model_id is not ForceNew

The issue's attribute table marks `model_id` as ForceNew, but the drift-handling section says "resolved model_id mismatch shown as diff → plan shows **update**" (not replace). An Update operation (in-place PUT with `reassign=true`) is more efficient and user-friendly than destroy+create, and is exactly what the `reassign` flag is designed for. This change documents `model_id` as updateable.

*Alternative considered:* Keep model_id ForceNew as listed in the attribute table. Rejected — it contradicts the explicit drift-handling requirement ("plan shows update") and wastes the reassign capability.

### 2. model_alias is the entitycore resource identity (GetResourceID)

The `model_alias` is the unique, stable name for the alias (must be unique across all models). It maps to `GetResourceID()` on the TF model, so the entitycore envelope uses it as the read/delete identity. The composite state `id` is `<cluster_uuid>/<model_alias>`.

### 3. Read uses GetTrainedModels with alias as model_id

The Elasticsearch typed client does not expose a dedicated "get alias" endpoint. Read calls `typedClient.Ml.GetTrainedModels().ModelId(alias).Do(ctx)`. An empty result or 404 means the alias does not exist (not-found). The resolved `model_id` is extracted from the returned `TrainedModelConfig`.

### 4. Delete resolves current model_id by alias first

The DELETE API path includes the model_id: `DELETE /_ml/trained_models/{model_id}/model_aliases/{model_alias}`. To avoid orphaning an alias when state is stale (e.g. the alias was reassigned out-of-band), the delete callback first calls `GET /_ml/trained_models/{model_alias}` to resolve the current model_id. If the alias no longer exists (404 or empty result), Delete treats it as already-gone. Otherwise, it calls DELETE with the resolved model_id. A 404 from the DELETE call is treated as idempotent success.

### 5. reassign flag semantics

`reassign` is an optional boolean (default false). It is sent as a query parameter on every PUT (create and update). When false, the PUT fails if the alias already points to a different model. When true, the PUT succeeds. In update scenarios, `reassign=true` is required to change `model_id`. The spec will document this constraint.

### 6. Client wrappers in internal/clients/elasticsearch/

Following repo convention, API calls are wrapped in `internal/clients/elasticsearch/ml_trained_model_alias.go`:
- `PutMLTrainedModelAlias(ctx, client, modelID, alias, reassign)` — Create/Update
- `GetMLTrainedModelAlias(ctx, client, alias)` — Read (returns model_id or not-found)
- `DeleteMLTrainedModelAlias(ctx, client, modelID, alias)` — Delete

### 7. Acceptance test for "fail without reassign=true"

Test scenario 3 ("Reassign without `reassign = true` → should fail") and scenario 7 ("Alias collision without reassign → should fail") test error conditions. These can be implemented as negative tests using `resource.TestExpectError` or as step-level `ExpectError` checks.

## Open questions

None — the issue body provides sufficient detail for implementation.
