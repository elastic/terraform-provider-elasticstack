## Why

The Fleet agentless policies API (`/api/fleet/agentless_policies`) is being deprecated in favour of the renamed, cleaner surface `/api/fleet/managed_integrations`. The new API introduces full CRUD (POST/GET/PUT/DELETE) including a dedicated GET and PUT that return a clean `KibanaHTTPAPIsManagedIntegration` response type, eliminating the leaky `PackagePolicy` fallback used today. The product has been renamed from "agentless" to "Elastic Managed Integration".

The existing `elasticstack_fleet_agentless_policy` resource targets the deprecated surface with fallback read/update paths through `/api/fleet/package_policies/{id}`. A `models_convert.go` of ~847 lines exists mainly to filter leaky `PackagePolicy` internals out of state. Migrating to the new API removes that complexity, gains in-place `name` and `package.version` updates, and aligns the provider with the going-forward Elastic endpoint.

## What Changes

- Remove the `elasticstack_fleet_agentless_policy` resource (no deprecation shim; no state migration or state upgrade function).
- Add `elasticstack_fleet_managed_integration` resource backed by the `managed_integrations` CRUD endpoints. The resource is registered in `experimentalResources()`, matching the upstream tech-preview status.
- The client layer moves from the deprecated `agentless_policies` endpoints (with package_policy fallbacks) to the dedicated `managed_integrations` CRUD surface: `POST`, `GET/{id}`, `PUT/{id}`, `DELETE/{id}`.
- Schema changes: `name` and `package.version` become updatable in-place (drop `RequiresReplace`); `package.name` stays `RequiresReplace` (immutable upstream). `global_data_tags` is remodelled from `ListNestedAttribute{name, value:string}` to `MapNestedAttribute` keyed by tag name with `{string_value, number_value}` (aligning with `elasticstack_fleet_agent_policy`).
- The read/update code is simplified significantly: `populateFromPackagePolicy` is replaced by `populateFromManagedIntegration` (direct read of the clean response), and `update.go`'s echo-current/overlay machinery is replaced by a full-replace body built from the plan.
- The version gate in `capabilities.go` is moved from the 9.3.0 `agentless_policies` floor to 9.5.0, the Kibana version that introduced `/api/fleet/managed_integrations` (verified against a 9.5.0-SNAPSHOT build; the same version already used as `policyshape.MinVersionCondition`). Because the new floor now equals the `condition`-support version, the separate `condition` capability check is removed as redundant.
- `_upgrade` / `_upgrade/dryrun` endpoints are deferred to a separate change.

## Capabilities

### New Capabilities

- `fleet-managed-integration`: Defines the schema and runtime behaviour of the `elasticstack_fleet_managed_integration` resource, including full CRUD against the managed_integrations API, in-place name/version update, the full-replace PUT body semantics, and the `global_data_tags` MapNestedAttribute modelling.

### Modified Capabilities

- `fleet-agentless-policy`: REMOVED. The `elasticstack_fleet_agentless_policy` resource is deleted; this capability is superseded by `fleet-managed-integration`.

## Impact

- **New code**: `internal/fleet/managedintegration/` (resource), `internal/clients/fleet/managed_integration.go` (thin client wrappers).
- **Deleted code**: `internal/fleet/agentlesspolicy/` (entire package), `internal/clients/fleet/agentless_policy.go` (and test).
- **New docs/examples**: `examples/resources/elasticstack_fleet_managed_integration/`; generated `docs/resources/fleet_managed_integration.md`.
- **Deleted docs/examples**: `examples/resources/elasticstack_fleet_agentless_policy/`, `docs/resources/fleet_agentless_policy.md`.
- **Provider registration**: Remove `agentlesspolicy.NewResource` from `experimentalResources()`; add `managedintegration.NewResource` there.
- **Generated clients**: `generated/kbapi/kibana.gen.go` already contains the `managed_integrations` CRUD surface (`Post/Get/Put/DeleteFleetManagedIntegrations...WithResponse`); no regeneration needed.
- **No released-user impact**: `elasticstack_fleet_agentless_policy` has never shipped in a release (still listed under `## [Unreleased]` in `CHANGELOG.md`); this is a pre-release rename, not a breaking change for practitioners. No migration guide, state-mv/re-import instructions, or CHANGELOG "Breaking changes" entry are needed.
