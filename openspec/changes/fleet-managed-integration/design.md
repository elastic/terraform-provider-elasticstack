## Context

The existing `elasticstack_fleet_agentless_policy` resource was built against the `/api/fleet/agentless_policies` surface, which only offered CREATE and DELETE. Read and Update fell back through `/api/fleet/package_policies/{id}`, returning a full `PackagePolicy` type that leaks Fleet internals (`policy_ids`, `revision`, `secret_references`, `version`, `output_id`, `supports_agentless`, top-level `enabled`). The `models_convert.go` file (~847 lines) exists primarily to filter those internals out of state. The new `/api/fleet/managed_integrations` surface exposes full CRUD with a clean `KibanaHTTPAPIsManagedIntegration` response type that maps almost exactly to the existing schema, so the migration dramatically simplifies the conversion layer.

The resource package is moved wholesale (`internal/fleet/agentlesspolicy/` → `internal/fleet/managedintegration/`) with targeted changes to the API client calls, schema modifiers, and the two API-coupled files (`models_convert.go`, `update.go`). Topology check, capability versioning, package cache lookup, and the entity-core wiring carry over unchanged.

## Goals / Non-Goals

**Goals:**
- Replace all four client calls with their `managed_integrations` equivalents; remove the `package_policies` fallback paths.
- Simplify `models_convert.go` by replacing `populateFromPackagePolicy` with `populateFromManagedIntegration` that reads the clean response directly.
- Simplify `update.go` by replacing the echo-current/overlay pattern with a full-replace body built from plan.
- Drop `RequiresReplace` from `name` and `package.version` (now updatable in-place).
- Remodel `global_data_tags` as `MapNestedAttribute{name → {string_value, number_value}}` (aligning with `elasticstack_fleet_agent_policy`).
- Update the version gate to the Kibana release that introduced `/api/fleet/managed_integrations`.

**Non-Goals:**
- `_upgrade` / `_upgrade/dryrun` bulk endpoints — deferred to a separate change.
- Per-stream `var_group_selections` — optional additive follow-up.
- `created_by` / `updated_by` Computed fields — optional additive follow-up.
- State migration, state upgrade functions, or deprecation shims for the old resource type.
- Retrofitting other Fleet resources or shared utilities.

## Decisions

### Decision 1: Resource naming — new type, no deprecation shim

New resource type `elasticstack_fleet_managed_integration`; old type `elasticstack_fleet_agentless_policy` removed entirely with no compatibility shim or deprecation warning.

**Why:** The issue explicitly decided migrate-and-rename with no shim. Keeping both types would require maintaining both client paths indefinitely. Users migrate via `terraform state mv` (structurally unchanged attributes) or re-import.

### Decision 2: Move-and-rename, not rewrite

Copy `internal/fleet/agentlesspolicy/` → `internal/fleet/managedintegration/`; rename the Go `package` declaration; update all references. Endpoint-agnostic files (`topology.go`, `capabilities.go`, `package_cache.go`, `resource.go`, `models.go`, test scaffolding) carry over with mechanical renames only.

**Why:** The underlying schema and test patterns are sound. A rewrite risks regressions and loses the existing test coverage that validates correct behaviour. Move-and-rename isolates exactly the changes that matter: the four client calls and the two API-coupled files.

### Decision 3: No state migration

Users transition via `terraform state mv` or `terraform import`. No `StateUpgraders` or migration helper.

**Why:** The schema changes (`name`/`package.version` mutability, `global_data_tags` shape) make a fully automatic migration non-trivial and the issue explicitly accepted this trade-off.

### Decision 4: `name` and `package.version` become updatable

Drop `RequiresReplace` from `name` and `package.version` in `schema.go`; keep `RequiresReplace` on `package.name` (immutable upstream). The PUT body includes both from the plan.

**Why:** The new API's PUT supports updating both fields in-place. Dropping `RequiresReplace` exposes this capability without a breaking change to users who were previously forced to recreate resources on name changes.

### Decision 5: `global_data_tags` shape change

Rewrite from `ListNestedAttribute{name, value:string}` to `MapNestedAttribute` keyed by tag name, with item `{string_value: StringAttribute, number_value: Float32Attribute}`, `ConflictsWith`+`AtLeastOneOf` validators. Mirror `internal/fleet/agentpolicy/schema.go`.

**Why:** The old shape was an outlier and stringified numbers. Aligning with the sibling `elasticstack_fleet_agent_policy` resource simplifies cross-resource usage. The structural change means `state mv` is not clean for this block; accepted under the no-migration stance.

### Decision 6: `cloud_connector` handling unchanged

Keep `SingleNestedAttribute`, all sub-fields `RequiresReplace`, `name`/`target_csp` preserved from state on Read. On PUT (full-replace), `{enabled, cloud_connector_id}` is derived from state and always re-sent — omitting it would detach the connector. `name`/`target_csp` are never sent on PUT (write-only fields that don't round-trip).

**Why:** Full-replace semantics require re-sending `cloud_connector` on every update to avoid accidental connector detachment. Preserving `name`/`target_csp` from state on Read is the existing pattern; the new clean response type also does not return these fields.

### Decision 7: `update.go` full-replace simplification

Remove the echo-current/overlay machinery. `buildUpdateBody` takes only the plan (desired state); no `*kbapi.PackagePolicy` "current" parameter. Build `KibanaHTTPAPIsCreateManagedIntegrationRequest` directly from plan using the same `decodeInputs`+`applyCreateInputs` helpers the create path uses.

**Why:** The new PUT is full-replace. The old echo-current pattern existed solely to work around the lack of a dedicated PUT. `mergeVarsInto` is likely removable as full-replace sends plan vars wholesale.

### Decision 8: Version gate update (the critical spike)

Move the `EnforceMinVersion` floor from 9.3.0 to the Kibana version that introduced `/api/fleet/managed_integrations` (landed in kibana#276925). The 9.5.0 `condition` gate stays as-is. The exact version must be confirmed against that PR or the shipped OAS changelog before setting the constant.

**Why:** Using the wrong floor causes 404s against stacks that have `agentless_policies` but not `managed_integrations`. This is the highest-risk item in the migration and must be resolved before implementation is considered complete.

### Decision 9: `experimentalResources()` placement unchanged

Register the new resource in `experimentalResources()`, matching the upstream tech-preview status.

## Open Questions

1. **Exact Kibana version for `/api/fleet/managed_integrations`**: Must be confirmed against kibana#276925 or the shipped OAS changelog. Getting this wrong 404s against supported-but-older stacks. This is the highest-priority spike.
2. **`onlyCreateOnlyFlagsChanged` short-circuit**: Re-evaluate whether this optimisation still applies under full-replace PUT semantics. May simplify to a "no-op if full desired-state body equals last-applied body" check, or may be removable entirely.
3. **Per-stream `var_group_selections`**: Currently only top-level is supported. Confirm whether the new API's stream-level field should be exposed now or deferred.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Wrong version gate** 404s against stacks with `agentless_policies` but not `managed_integrations` | Resolve the version spike first; add acceptance test skip-gated to the new floor |
| **Full-replace PUT missing `cloud_connector`** detaches connectors or clears optional fields | Unit + acceptance tests covering cloud_connector update path; explicitly test that cloud_connector is re-sent from state |
| **`state mv` incompatibility for `global_data_tags`** confuses migrating users | Document the shape change in CHANGELOG and resource description; note that re-import is required for this block |
| **`package_policies` fallback removal** breaks users on older Kibana versions | Enforced by version gate; the old fallback only existed because the new endpoint didn't exist |
