# Implementation notes — task 1 (pre-implementation)

Artifacts for OpenSpec change `fleet-managed-integration`, section **1. Pre-implementation**.

## Intermediate branch state (after tasks 7–8)

Task 3 moved the resource package to `internal/fleet/managedintegration/` and registered **`elasticstack_fleet_managed_integration`** in `experimentalResources()`.

Tasks 4–6 completed version gate, schema, and `models_convert.go` simplification against `KibanaHTTPAPIsManagedIntegration`.

**Task 7 (complete)** rewrote `update.go` for managed_integrations **full-replace PUT**:

- `buildUpdateBody(plan, prior)` compiles `PutFleetManagedIntegrationsPolicyidJSONRequestBody` from the plan via shared `toManagedIntegrationRequestBody`; **`prior` supplies `cloud_connector {enabled, cloud_connector_id}` only** (never `name`/`target_csp`).
- Update **PUT** calls `UpdateManagedIntegration`; the write callback returns **plan only** — no `populateFromManagedIntegration` on the PUT response. Final state comes from the entitycore envelope **Read-after-write** (coding-standards.md).
- **`onlyCreateOnlyFlagsChanged`** skips Fleet calls and sets **`KibanaWriteResult.SkipReadAfterWrite`** so the envelope persists plan without Read or PostRead (no managed_integrations GET/PUT/DELETE). The write callback merges known server-computed fields from prior state (for example `updated_at`) into the returned model when the plan leaves them Unknown.
- Full-replace optional fields: **known-null → omit** (generated `omitempty` clears on API); **unknown top-level API-backed optionals → attribute error**; known-empty collections sent explicitly where `sendExplicitEmptyScalars` applies.

**Task 8 (complete)** rewired Create/Read/Delete to `CreateManagedIntegration` / `ReadManagedIntegration` / `DeleteManagedIntegration`; removed `agentless_policy_compat.go` and `package_policy_read_bridge.go`. Create returns plan with server-assigned id only; final state comes from envelope read-after-write (same as Update).

Acceptance fixtures use `elasticstack_fleet_managed_integration`; test renames remain tasks 10–11. Live in-place **name**, **package.version**, and **cloud_connector** persistence are tracked in tasks **11.3–11.4** and **11.7** (not yet implemented).

## ~~Temporary schema vs update-body mismatch~~ (closed in task 7)

Task 5.1/5.2 made `name` and `package.version` updatable in Terraform; task 7.3 includes both in every full-replace PUT body from plan. Acceptance merge sign-off for live rename/version bump still requires tasks **11.3–11.4**.

## 1.1 MinVersion floor — **9.5.0**

**Decision:** Set `MinVersion` to `9.5.0` in `internal/fleet/managedintegration/models.go` (constant). Task 4.2 removed the separate condition capability gate; there is no `capabilities.go` in this package anymore.

**Rationale:**

- `/api/fleet/managed_integrations` was verified on a **9.5.0-SNAPSHOT** Kibana build (see `design.md` Decision 8).
- This matches `policyshape.MinVersionCondition` (`9.5.0`), so a separate `SupportsCondition` runtime gate is redundant (removed in task 4.2).
- Using `9.3.0` (the old `agentless_policies` floor) would allow plans against stacks that have the deprecated surface but not `managed_integrations`, producing 404s.

**Code touchpoints (task 1.1 / task 4):**

- `internal/fleet/managedintegration/models.go` — `MinVersion`, `GetVersionRequirements` error text/comments; `TestMinVersion_matchesPolicyshapeMinVersionCondition` guards alignment with `policyshape.MinVersionCondition`.

## 1.2 `KibanaHTTPAPIsManagedIntegration` ↔ schema mapping

Reviewed `generated/kbapi/kibana.gen.go` (`KibanaHTTPAPIsManagedIntegration`, lines ~50017–50091) against `internal/fleet/managedintegration/schema.go` and `models.go`.

### Direct mappings (no conversion surprise)

| Terraform attribute | API field | Notes |
|---------------------|-----------|-------|
| `policy_id`, `id` | `id` | Composite ID built as `<space_id>/<id>` |
| `name` | `name` | Updatable in-place on PUT (task 5.1) |
| `description` | `description` | Optional pointer; empty string folded to null |
| `namespace` | `namespace` | Optional pointer |
| `package.name` | `package.name` | Immutable upstream (`RequiresReplace` retained) |
| `package.version` | `package.version` | Updatable in-place on PUT (task 5.2) |
| `package.title` | `package.title` | Computed from registry when omitted |
| `vars_json` | `vars` | Typed union vars map |
| `var_group_selections` | `var_group_selections` | Top-level map only |
| `inputs` | `inputs` | Map keyed by `"<policy_template>-<input_type>"` (see `mappedInputKey`; bare `input_type` when policy template is empty) |
| `cloud_connector.enabled` | `cloud_connector.enabled` | |
| `cloud_connector.cloud_connector_id` | `cloud_connector.cloud_connector_id` | |
| `additional_datastreams_permissions` | `additional_datastreams_permissions` | List in schema, `*[]string` in API |
| `created_at` | `created_at` | |
| `updated_at` | `updated_at` | |

### Expected non-round-trip (preserve from prior state on Read)

| Terraform attribute | API GET | Handling |
|---------------------|---------|----------|
| `space_ids` | absent | Default from request space / import composite ID |
| `policy_template` | absent | Create request only; preserve from config/state |
| `force` | absent | Create-only |
| `create_dataset_templates` | absent | Create-only |
| `skip_topology_check` | absent | Create-only (Create preflight only) |
| `force_delete` | absent | Delete query param only |
| `cloud_connector.name` | absent | Write-only; preserve from state |
| `cloud_connector.target_csp` | absent | Write-only; preserve from state |

### Schema shape discrepancies (addressed in later tasks)

| Topic | API | Schema (post task 5) | Follow-up |
|-------|-----|----------------------|-----------|
| `global_data_tags` | `[{name, value: string\|number}]` | `MapNestedAttribute` with `string_value` / `number_value` | Task 6.4 conversion cleanup; task 11.5 live `number_value` case |
| Per-stream `var_group_selections` | present on `inputs.*.streams.*` | not modeled | Deferred (design.md non-goal) |
| `created_by` / `updated_by` | present on response | not in schema | Deferred (design.md non-goal) |
| `inputs.*.deprecated` / stream deprecation | present | not in schema | Intentionally ignored (Fleet internal metadata) |

### PackagePolicy leakage — eliminated

`KibanaHTTPAPIsManagedIntegration` does **not** expose `policy_ids`, `revision`, `secret_references`, `output_id`, `supports_agentless`, or top-level `enabled`. Task 6 replaces `populateFromPackagePolicy` with `populateFromManagedIntegration` so Read no longer filters PackagePolicy internals.

## 1.3 `onlyCreateOnlyFlagsChanged` under full-replace PUT

**Decision:** **Retain** the short-circuit unchanged in spirit; re-wire it in the rewritten `update.go` (task 7.6).

**Analysis:**

- `create_dataset_templates`, `force`, `force_delete`, and `skip_topology_check` remain **outside** the PUT body under full-replace semantics (same as today).
- None are `RequiresReplace`, so Terraform still invokes `Update` when only they change.
- Sending a full-replace PUT when the diff is confined to these flags would either be a no-op at best or risk unintended side effects; the delta spec requirement "Create/delete-only flag updates skip managed_integrations API calls" requires **no managed_integrations GET/PUT/DELETE** for such changes.
- The existing comparison (all API-backed fields equal, excluding provider plumbing and computed timestamps) remains correct once `name` and `package.version` become updatable — both are already in the comparison chain.
- Full-replace does **not** simplify this to a body-equality check: the plan never includes server-only timestamps, and building the PUT body just to compare would still require a Read.

**Implementation note for task 7:** Keep `onlyCreateOnlyFlagsChanged`; set `KibanaWriteResult.SkipReadAfterWrite` on that path so the envelope does not invoke Read (no managed_integrations GET). Normal Update still read-after-writes via the Read callback after PUT.

## 2. New managed_integrations client (task 2 / task 8)

Task 2 adds `internal/clients/fleet/managed_integration.go` with CRUD wrappers targeting `/api/fleet/managed_integrations`. The deprecated `agentless_policy.go` client file was removed. Task 8 rewired the resource package to call those wrappers exclusively; the temporary `agentless_policy_compat.go` bridge was removed.
