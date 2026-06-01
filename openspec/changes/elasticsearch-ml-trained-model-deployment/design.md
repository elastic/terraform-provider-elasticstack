## Context

Elasticsearch exposes trained model deployment lifecycle through three API endpoints:

- `POST _ml/trained_models/{model_id}/deployment/_start` — deploys (starts) a model; accepts allocation parameters, priority, and queue settings.
- `GET _ml/trained_models/{model_id}/_stats` — returns deployment stats including `deployment_stats.state`, `deployment_stats.allocation_status`, and per-allocation detail.
- `POST _ml/trained_models/{deployment_id}/deployment/_update` — updates mutable deployment parameters (`number_of_allocations`, `adaptive_allocations`).
- `POST _ml/trained_models/{deployment_id}/deployment/_stop` — stops (undeploys) a model.

A `deployment_id` defaults to the `model_id` when not provided by the caller. The same model can have multiple deployments with different `deployment_id` values.

The typed Go client (`go-elasticsearch` v8) exposes:
- `client.Ml.StartTrainedModelDeployment(modelID)` → builder with `.Request(req)`, `.DeploymentId(...)`, `.NumberOfAllocations(...)`, `.Priority(...)`, `.QueueCapacity(...)`, `.ThreadsPerAllocation(...)`, `.Timeout(...)`, `.WaitFor(...)`, `.Do(ctx)`
- `client.Ml.GetTrainedModelsStats().ModelId(modelID).Do(ctx)` → `gettrainedmodelsstats.Response`
- `client.Ml.UpdateTrainedModelDeployment(deploymentID).Request(req).Do(ctx)`
- `client.Ml.StopTrainedModelDeployment(modelID).Force(...).Do(ctx)`

The `types.AdaptiveAllocationsSettings` struct has `Enabled bool`, `MaxNumberOfAllocations *int`, `MinNumberOfAllocations *int`.

## Goals

- Manage the full deployment lifecycle for Elasticsearch ML trained models through Terraform.
- Allow practitioners to control deployment parameters: allocations, threads, priority, queue capacity, and adaptive allocations.
- Expose computed state (deployment state, allocation status, raw stats JSON) for observability.
- Handle drift from adaptive allocations gracefully—suppress allocation diffs when the server controls allocation counts.
- Follow existing state-transition resource patterns for consistency and maintainability.

## Non-Goals

- Creating or uploading the underlying trained model (out of scope; use Eland or `elasticstack_elasticsearch_ml_trained_model` if/when it exists).
- Managing inference endpoints (separate resource / API).
- Supporting multiple deployments of the same model in a single resource invocation.

## Decisions

| Topic | Decision |
|-------|----------|
| Resource pattern | State-transition resource via `entitycore.NewElasticsearchResource[T]` with `PlaceholderElasticsearchWriteCallback` for Create/Update, overriding both on the concrete struct. Mirrors `internal/elasticsearch/ml/jobstate/resource.go`. |
| Package location | `internal/elasticsearch/ml/trainedmodeldeployment/` |
| File structure | `resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `acc_test.go`, `descriptions.go` |
| Identity | Composite `id` = `<cluster_uuid>/<deployment_id>`. `deployment_id` ForceNew, defaults to `model_id` at creation when not specified. `UseStateForUnknown` on both `id` and `deployment_id`. |
| ForceNew attributes | `model_id`, `deployment_id`, `threads_per_allocation`, `priority`, `queue_capacity`, `wait_for`, `api_timeout` — changes require stop-then-start (destroy + create). |
| Mutable attributes | `number_of_allocations`, `adaptive_allocations` block — use Update API. |
| Adaptive allocations drift | When `adaptive_allocations.enabled = true`, the server controls effective `number_of_allocations`. Add a plan modifier (`ModifyPlan`) that suppresses the diff on `number_of_allocations` when adaptive allocations are enabled, to prevent perpetual plan noise. |
| `number_of_allocations` on read | When `adaptive_allocations.enabled = true`, do not overwrite state's `number_of_allocations` from the API on Read (preserve user intent). When `adaptive_allocations.enabled = false`, update from API. |
| Wait-for polling | After Start, if `wait_for` is set, poll `_stats` until the allocation status matches or the configured timeout elapses. Reuse `internal/asyncutils/state_waiter.go` pattern. |
| `stats_json` | Populated on every Read as the raw JSON of `TrainedModelStats` from the API. Read-only from the practitioner's perspective; useful for debugging or extracting fields not yet modelled. |
| `state` and `allocation_status` | Computed strings derived from `deployment_stats.state` and `deployment_stats.allocation_status.state` in the stats response. |
| Delete (stop) | Call Stop Deployment API. Treat HTTP 404 as success (idempotent). Force-stop on timeout if needed. |
| External stop (drift) | If the deployment is stopped externally, Read will see state = "stopped". The next plan shows a diff and the next apply calls Start again. This is expected state-transition behaviour. |
| Client wrappers | Add functions in `internal/clients/elasticsearch/`: `StartTrainedModelDeployment`, `GetTrainedModelStats`, `UpdateTrainedModelDeployment`, `StopTrainedModelDeployment`. |
| Provider registration | Add `NewTrainedModelDeploymentResource()` to `resources()` list in `provider/plugin_framework.go`. |
| Minimum ES version | Trained model deployment APIs are GA in Elasticsearch 8.0+. No version gate required beyond the provider's existing minimum ES version. |

## Non-Goals (implementation)

- Do not add support for inferencing or calling the inference API from within this resource.
- Do not model every field in `TrainedModelStats`; `stats_json` covers extensibility.

## Risks / Trade-offs

- **Adaptive allocations conflict with `number_of_allocations`**: The plan modifier approach is a heuristic—if the user explicitly sets `number_of_allocations` AND enables adaptive allocations, the intent is ambiguous. The spec requires surfacing a validation error in this case.
- **Start-API timeout vs Terraform timeout**: `api_timeout` is the server-side deployment start timeout. The Terraform `timeouts.create` covers the total wait including post-start polling. These are orthogonal; document clearly.
- **Multiple deployments per model**: The Elasticsearch API allows deploying the same model with different `deployment_id` values. This resource manages exactly one deployment. Practitioners who need multiple deployments create multiple resource instances with distinct `deployment_id` values.
- **Import with model_id vs deployment_id**: The composite ID format `<cluster_uuid>/<deployment_id>` is used for import. Since `deployment_id` defaults to `model_id`, import works naturally for the common case.

## Open Questions

1. **`number_of_allocations` validation**: Should the provider return an error at plan time when `adaptive_allocations.enabled = true` and `number_of_allocations` is also explicitly set? Or should it silently ignore? The conservative choice (emit a diagnostic warning) is noted here; implementation MUST confirm the UX.
2. **`wait_for` default**: Should `wait_for` default to `"started"` if omitted, or leave the choice to the Elasticsearch server default? Confirm against API docs during implementation.
3. **Stop force flag**: Should the Delete path always pass `force=true` to the Stop API, or should force be a configurable attribute? The issue body leaves this unspecified; default to `force=false` with documentation noting it can be destroyed with force if needed. This SHOULD be resolved during implementation.
4. **Stats polling on Read**: `GET _ml/trained_models/{model_id}/_stats` returns stats per `deployment_id`. Confirm that querying by `model_id` returns stats for all deployments of that model, and filter by `deployment_id` client-side if needed.

## Migration / State

- This is a new resource; no state migration required.
- Import is supported via `ImportStatePassthroughID` on `id`.
