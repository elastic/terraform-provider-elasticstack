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
- Update the version gate to the Kibana release that introduced `/api/fleet/managed_integrations` (confirmed: 9.5.0 — see Decision 8).
- Simplify away the separate `condition`-support capability check, now redundant with the resource-level version floor (see Decision 8).

**Non-Goals:**
- `_upgrade` / `_upgrade/dryrun` bulk endpoints — deferred to a separate change.
- Per-stream `var_group_selections` — optional additive follow-up.
- `created_by` / `updated_by` Computed fields — optional additive follow-up.
- State migration, state upgrade functions, or deprecation shims for the old resource type.
- Retrofitting other Fleet resources or shared utilities.

## Decisions

### Decision 1: Resource naming — new type, no deprecation shim

New resource type `elasticstack_fleet_managed_integration`; old type `elasticstack_fleet_agentless_policy` removed entirely with no compatibility shim or deprecation warning.

**Why:** The issue explicitly decided migrate-and-rename with no shim. Keeping both types would require maintaining both client paths indefinitely. `elasticstack_fleet_agentless_policy` has never shipped in a release (it is still listed under `## [Unreleased]` in `CHANGELOG.md`), so there are no released users to provide migration guidance for; this is a pre-release rename, not a breaking change in practice.

### Decision 2: Move-and-rename, not rewrite

Copy `internal/fleet/agentlesspolicy/` → `internal/fleet/managedintegration/`; rename the Go `package` declaration; update all references. Endpoint-agnostic files (`topology.go`, `capabilities.go`, `package_cache.go`, `resource.go`, `models.go`, test scaffolding) carry over with mechanical renames only.

**Why:** The underlying schema and test patterns are sound. A rewrite risks regressions and loses the existing test coverage that validates correct behaviour. Move-and-rename isolates exactly the changes that matter: the four client calls and the two API-coupled files.

### Decision 3: No state migration

No `StateUpgraders` or migration helper. Since `elasticstack_fleet_agentless_policy` has never shipped in a release, there is no released-user population to write migration documentation for; anyone tracking `main` directly can use `terraform state mv` for structurally-unchanged attributes or re-import, but this is not a supported/documented migration path.

**Why:** The schema changes (`name`/`package.version` mutability, `global_data_tags` shape) make a fully automatic migration non-trivial, and the issue explicitly accepted this trade-off. Because the old resource is unreleased, the trade-off costs nothing in practice.

### Decision 4: `name` and `package.version` become updatable

Drop `RequiresReplace` from `name` and `package.version` in `schema.go`; keep `RequiresReplace` on `package.name` (immutable upstream). The PUT body includes both from the plan.

**Why:** The new API's PUT supports updating both fields in-place. Dropping `RequiresReplace` exposes this capability without a breaking change to users who were previously forced to recreate resources on name changes.

### Decision 5: `global_data_tags` shape change

Rewrite from `ListNestedAttribute{name, value:string}` to `MapNestedAttribute` keyed by tag name, with item `{string_value: StringAttribute, number_value: Float32Attribute}`, `ConflictsWith`+`AtLeastOneOf` validators. Mirror `internal/fleet/agentpolicy/schema.go`.

**Why:** The old shape was an outlier and stringified numbers. Aligning with the sibling `elasticstack_fleet_agent_policy` resource simplifies cross-resource usage. The structural change would make `state mv` non-clean for this block, but since the resource is unreleased this has no practical migration cost.

### Decision 6: `cloud_connector` handling unchanged

Keep `SingleNestedAttribute` with a single object-level `RequiresReplace` plan modifier (not one per sub-field) — this already forces replacement when any sub-field changes. `name`/`target_csp` preserved from state on Read. On PUT (full-replace), `{enabled, cloud_connector_id}` is derived from state and always re-sent — omitting it would detach the connector. `name`/`target_csp` are never sent on PUT (write-only fields that don't round-trip).

**Why:** Full-replace semantics require re-sending `cloud_connector` on every update to avoid accidental connector detachment. Preserving `name`/`target_csp` from state on Read is the existing pattern; the new clean response type also does not return these fields.

### Decision 7: `update.go` full-replace simplification

Remove the echo-current/overlay machinery. `buildUpdateBody` takes only the plan (desired state); no `*kbapi.PackagePolicy` "current" parameter. Build `KibanaHTTPAPIsCreateManagedIntegrationRequest` directly from plan using the same `decodeInputs`+`applyCreateInputs` helpers the create path uses.

**Why:** The new PUT is full-replace. The old echo-current pattern existed solely to work around the lack of a dedicated PUT. `mergeVarsInto` is likely removable as full-replace sends plan vars wholesale.

### Decision 8: Version gate update — confirmed at 9.5.0, condition gate simplified away

Move the `EnforceMinVersion` floor from 9.3.0 to **9.5.0** (verified against a 9.5.0-SNAPSHOT Kibana build; the same version already used as `policyshape.MinVersionCondition`). This is now a resource-level `MinVersion` constant in `models.go`/`capabilities.go`, mirroring the pattern in `internal/fleet/agentlesspolicy/models.go`.

Shared client version checks (`internal/clients/version_utils.go`) treat a same-core **`-SNAPSHOT`** server build as satisfying a **release** minimum (e.g. Kibana `9.5.0-SNAPSHOT` meets floor `9.5.0`), matching CI matrix stacks and acceptance tests that probe Kibana via `EnforceMinVersion` rather than Elasticsearch alone.

Because the new floor is identical to `MinVersionCondition`, the separate per-request `condition`-support capability check (`agentlessPolicyFeatures.SupportsCondition`, `resolveAgentlessPolicyFeatures`, and `validateInputConditionSupport` in `models_convert.go`) is now redundant: a stack that can run this resource at all is guaranteed to support `condition`. That capability check, and its dedicated gating, is removed; `condition` is treated as unconditionally supported once the resource-level floor is satisfied.

**Why:** Using the wrong floor causes 404s against stacks that have `agentless_policies` but not `managed_integrations` — this was the highest-risk item in the migration and is now resolved. Collapsing the two gates removes a runtime check that can no longer produce a different answer than the resource-level floor, simplifying `capabilities.go` and `models_convert.go` without losing any protection.

### Decision 9: `experimentalResources()` placement unchanged

Register the new resource in `experimentalResources()`, matching the upstream tech-preview status.

## Open Questions

_(none — task 1.3 resolved below)_

### Decision 10: Retain `onlyCreateOnlyFlagsChanged` (task 1.3)

**Decision:** Keep the create/delete-only-flags short-circuit when rewriting `update.go` for full-replace PUT (task 7.6), satisfying the delta spec requirement "Create/delete-only flag updates skip managed_integrations API calls".

**Why:** `force`, `create_dataset_templates`, `force_delete`, and `skip_topology_check` remain outside the PUT body. Terraform still invokes Update when only they change. The short-circuit writes plan to state without any managed_integrations GET/PUT/DELETE call. Full-replace semantics do not make a body-equality check simpler (timestamps and write-only fields are not in the plan). See `implementation.md` §1.3.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Full-replace PUT missing `cloud_connector`** detaches connectors or clears optional fields | Unit + acceptance tests covering cloud_connector update path; explicitly test that cloud_connector is re-sent from state |
| **`package_policies` fallback removal** breaks users on older Kibana versions | Enforced by version gate (9.5.0); the old fallback only existed because the new endpoint didn't exist |
