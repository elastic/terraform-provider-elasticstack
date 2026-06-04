## Why

Three Kibana/Fleet resources — `fleet_integration`, `kibana_synthetics_monitor`, and `kibana_slo` — still embed only `*entitycore.ResourceBase` and hand-roll their own CRUD methods. Every other resource in the `kibana/` and `fleet/` trees has already migrated to the `entitycore.KibanaResource[T]` envelope, which centralises connection-block injection, space-ID validation, client resolution, and read-after-write. Completing the migration eliminates a class of divergence bugs and reduces the maintenance surface to a consistent pattern.

## What Changes

- **`internal/fleet/integration`**: Replace `*entitycore.ResourceBase` with `*entitycore.KibanaResource[integrationModel]`. Add `KibanaResourceModel` interface methods to `integrationModel` plus `IsUnscopedSpace()` (SpaceID is optional). Rewrite CRUD as envelope callbacks. Remove explicit `kibana_connection` block from schema (envelope injects it). Retain `ResourceWithUpgradeState` on the concrete type.
- **`internal/kibana/synthetics/monitor`**: Replace `*entitycore.ResourceBase` with `*entitycore.KibanaResource[tfModelV0]`. Add `KibanaResourceModel` interface methods (composite-ID-aware). Remove the dead `synthetics.ESAPIClient` interface, its sole implementation on the concrete type, and the `synthetics.GetKibanaOAPIClient(ESAPIClient, …)` helper — confirmed no callers outside the now-deleted definition and assertion. Rewrite CRUD as envelope callbacks.
- **`internal/kibana/slo`**: Replace `*entitycore.ResourceBase` with `*entitycore.KibanaResource[tfModel]`. Add `KibanaResourceModel` interface methods. Promote `readAndPopulate` from a method to a package-level function so write callbacks can call it. Move `reconcileSloEnabledAfterWrite` (the enable/disable reconciliation that requires an intermediate read because the create response only returns `id`) into the Create and Update write callbacks. Retain `ResourceWithConfigValidators` and `ResourceWithUpgradeState` on the concrete type.
- **`entitycore_contract_test.go`** added to each migrated package asserting `*entitycore.KibanaResource[T]` embedding.

## Capabilities

### New Capabilities

None — this change does not alter external behaviour.

### Modified Capabilities

None — existing specs for `fleet-integration`, `kibana-synthetics-monitor`, and `kibana-slo` are unchanged. The `entitycore-kibana-resource-envelope` spec is also unchanged.

## Impact

- `internal/fleet/integration/` — all non-test Go files touched
- `internal/kibana/synthetics/monitor/` — all non-test Go files touched; `synthetics/api_client.go` simplified
- `internal/kibana/slo/` — all non-test Go files touched
- No schema changes visible to practitioners; acceptance tests unaffected
- Dependent packages importing these packages are unaffected (public API unchanged)
