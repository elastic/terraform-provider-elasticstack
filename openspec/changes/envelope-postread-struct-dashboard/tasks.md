## 1. Kibana envelope PostRead struct change

- [x] 1.1 Introduce `KibanaPostReadRequest[T]` struct in `kibana_resource_envelope.go` with fields `Client *clients.KibanaScopedClient`, `Prior T`, `State T`, `Private any`
- [x] 1.2 Change `KibanaPostReadFunc[T]` type to `func(ctx context.Context, req KibanaPostReadRequest[T]) (T, diag.Diagnostics)`
- [x] 1.3 Update `runKibanaWrite`: populate `req.Prior` from the plan model; invoke PostRead after the read callback; set state from PostRead's returned model (not the read-callback model directly); skip state set if PostRead returns error diagnostics
- [x] 1.4 Update `runKibanaRead`: populate `req.Prior` from the state model decoded before the read callback runs; invoke PostRead after the read callback; set state from PostRead's returned model; skip state set if PostRead returns error diagnostics
- [x] 1.5 Update `kibana_resource_envelope_test.go`: update all 7 existing PostRead test function lambdas to the new `KibanaPostReadRequest[T]` signature; add scenarios verifying `req.Prior` is the plan on write path, `req.Prior` is the prior state on read path, and state is set from PostRead's returned model

## 2. Elasticsearch envelope PostRead struct change

- [x] 2.1 Introduce `ElasticsearchPostReadRequest[T]` struct in `resource_envelope.go` with fields `Client *clients.ElasticsearchScopedClient`, `Prior T`, `State T`, `Private any`
- [x] 2.2 Change `PostReadFunc[T]` type to `func(ctx context.Context, req ElasticsearchPostReadRequest[T]) (T, diag.Diagnostics)`
- [x] 2.3 Update `runElasticsearchWrite` and `runElasticsearchRead` with the same treatment as the Kibana envelope in tasks 1.3–1.4
- [x] 2.4 Update `postReadPersistAPIKeyCapabilities` in `internal/elasticsearch/security/apikey/resource/resource.go` to the new signature; function ignores the model, so return `(req.State, diags)` unchanged
- [x] 2.5 Update `resource_envelope_test.go`: update all 7 existing PostRead test function lambdas to the new `ElasticsearchPostReadRequest[T]` signature; add the same `Prior` and state-set scenarios as task 1.5
- [x] 2.6 Confirm `make build` passes

## 3. kibana/dashboard migration

- [x] 3.1 Add `KibanaResourceModel` interface methods to `DashboardModel` (`GetID` returning composite `<space_id>/<dashboard_id>`, `GetResourceID` returning `DashboardID`, `GetSpaceID`, `GetKibanaConnection`)
- [x] 3.2 Replace `*entitycore.ResourceBase` embed with `*entitycore.KibanaResource[DashboardModel]` in `resource.go`
- [x] 3.3 Remove the explicit `"kibana_connection": providerschema.GetKbFWConnectionBlock()` entry from `schema.go` (envelope injects it; leaving it in is redundant and inconsistent with every other `KibanaResource[T]` consumer)
- [x] 3.4 Rewrite `create.go` as a `KibanaWriteFunc[DashboardModel]`: call Kibana create API, set `model.ID` (composite) and `model.DashboardID`, return `KibanaWriteResult[T]` — no alignment here
- [x] 3.5 Rewrite `update.go` as a `KibanaWriteFunc[DashboardModel]`: call Kibana update API, set `model.ID` and `model.DashboardID`, return `KibanaWriteResult[T]` — no alignment here
- [x] 3.6 Rewrite `read.go` as a `readDashboard(ctx, client, resourceID, spaceID, model) (DashboardModel, bool, diag.Diagnostics)` callback: raw API call and model population only — no alignment; envelope calls PostRead after this to apply alignment
- [x] 3.7 Rewrite `delete.go` as a `deleteDashboard(ctx, client, resourceID, spaceID, model) diag.Diagnostics` callback
- [x] 3.8 Implement `postReadDashboard` as a `KibanaPostReadFunc[DashboardModel]`: call all four alignment functions using `req.Prior` as the reference (plan on write path, prior state on read path):
  - `alignDashboardStateFromPlanPanels(req.Prior.Panels, req.State.Panels)`
  - `suppressReadTopLevelPanelsWhenPlanEmpty(req.Prior.Panels, &req.State)`
  - `alignDashboardStateFromPlanSections(ctx, req.Prior.Sections, req.State.Sections)`
  - `alignDashboardStateFromPlanPinnedPanels(ctx, req.Prior.PinnedPanels, req.State.PinnedPanels)`
  - return `(req.State, nil)`
- [x] 3.9 Wire all callbacks (including `PostRead: postReadDashboard`) into `entitycore.NewKibanaResource[DashboardModel]` in `resource.go`
- [x] 3.10 Update `entitycore_contract_test.go` (already exists asserting `*entitycore.ResourceBase`) to instead assert `*entitycore.KibanaResource[models.DashboardModel]` embedding

## 4. Final validation

- [x] 4.1 Run `make build` across the full provider
- [x] 4.2 Run dashboard acceptance tests to verify panel alignment behaviour is unchanged for create, update, and read paths
