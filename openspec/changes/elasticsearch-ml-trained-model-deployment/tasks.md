## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ml-trained-model-deployment --type change` (or `make check-openspec` after sync).
- [ ] 1.2 Resolve remaining open questions in `design.md` (force-stop flag, stats-query filtering by deployment_id); update delta spec with confirmed behaviour.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/elasticsearch-ml-trained-model-deployment/spec.md` or **archive** the change per project workflow.

## 2. Client Wrappers

- [ ] 2.1 Add `StartTrainedModelDeployment(ctx, apiClient, modelID string, req *starttrainedmodeldeployment.Request, opts ...) (*starttrainedmodeldeployment.Response, diag.Diagnostics)` to `internal/clients/elasticsearch/`.
- [ ] 2.2 Add `GetTrainedModelStats(ctx, apiClient, modelID string) (*types.TrainedModelStats, diag.Diagnostics)` to `internal/clients/elasticsearch/`. Filter by `deployment_id` if the API returns stats for all deployments of the model.
- [ ] 2.3 Add `UpdateTrainedModelDeployment(ctx, apiClient, deploymentID string, req *updatetrainedmodeldeployment.Request) diag.Diagnostics` to `internal/clients/elasticsearch/`.
- [ ] 2.4 Add `StopTrainedModelDeployment(ctx, apiClient, deploymentID string, force bool) diag.Diagnostics` to `internal/clients/elasticsearch/`. Treat HTTP 404 as success.

## 3. Package and Resource Implementation

- [ ] 3.1 Create package `internal/elasticsearch/ml/trainedmodeldeployment/`.
- [ ] 3.2 Implement `models.go`: define `TrainedModelDeploymentData` struct with `tfsdk` tags for all schema attributes (model_id, deployment_id, number_of_allocations, threads_per_allocation, priority, queue_capacity, wait_for, api_timeout, adaptive_allocations, id, state, allocation_status, stats_json, timeouts).
- [ ] 3.3 Implement `schema.go`: define `GetSchema()` returning the Plugin Framework schema. Mark ForceNew attributes with `RequiresReplace`. Add `ConflictsWith` validator so `number_of_allocations` and `adaptive_allocations` cannot both be configured.
- [ ] 3.4 Implement `resource.go`: register via `entitycore.NewElasticsearchResource[TrainedModelDeploymentData]` with `PlaceholderElasticsearchWriteCallback` for Create/Update; override `Create`, `Update` on the concrete struct; implement `ImportState` via `ImportStatePassthroughID`.
- [ ] 3.5 Implement `create.go`: call `StartTrainedModelDeployment` with user-specified parameters; poll `GetTrainedModelStats` until allocation status matches `wait_for` (default `"fully_allocated"`) or `timeouts.create` elapses (reuse `internal/asyncutils/state_waiter.go` pattern); populate computed attributes (id, deployment_id, state, allocation_status, stats_json).
- [ ] 3.6 Implement `read.go`: call `GetTrainedModelStats`; populate `state`, `allocation_status`, `stats_json`; update `number_of_allocations` from API when `adaptive_allocations` is not configured.
- [ ] 3.7 Implement `update.go`: call `UpdateTrainedModelDeployment` with updated `number_of_allocations` and/or `adaptive_allocations`; re-read stats; update state.
- [ ] 3.8 Implement `delete.go`: call `StopTrainedModelDeployment`; treat 404 as success; log warning if already stopped.
- [ ] 3.9 Implement `descriptions.go` (or `resource-description.md`): user-facing Markdown descriptions for schema attributes and resource overview.

## 4. Provider Registration

- [ ] 4.1 Add `trainedmodeldeployment.NewTrainedModelDeploymentResource()` to the `resources()` list in `provider/plugin_framework.go`.

## 5. Testing

- [ ] 5.1 Implement `acc_test.go`: skip all tests if no ML nodes are available or no suitable model exists (check for `skipFunc` pattern in sibling test files).
- [ ] 5.2 Add acceptance test: **Create and verify** — start a deployment on a pre-existing trained model; assert `state = "started"`, `deployment_id` is set, `id` matches `<cluster_uuid>/<deployment_id>`.
- [ ] 5.3 Add acceptance test: **Re-plan with no diff** — run `PlanOnly` after initial apply; assert no diff.
- [ ] 5.4 Add acceptance test: **Update `number_of_allocations`** — change value; assert update call made and state reflects new value.
- [ ] 5.5 Add acceptance test: **Update `adaptive_allocations`** — switch from fixed allocations to adaptive allocations; assert update applies adaptive settings and subsequent plan shows no diff.
- [ ] 5.6 Add acceptance test: **Import** — import existing deployment by composite id `<cluster_uuid>/<deployment_id>`; assert state matches.
- [ ] 5.7 Add acceptance test: **Delete** — destroy resource; assert deployment is stopped and resource removed from state.
- [ ] 5.8 Add acceptance test: **Model not found** — attempt to deploy a non-existent model id; assert Terraform error diagnostic.
- [ ] 5.9 Add unit tests for schema validation: assert `number_of_allocations` and `adaptive_allocations` cannot both be configured (ConflictsWith).
