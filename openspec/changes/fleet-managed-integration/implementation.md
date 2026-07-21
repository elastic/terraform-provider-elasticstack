# Implementation notes — task 1 (pre-implementation)

Artifacts for OpenSpec change `fleet-managed-integration`, section **1. Pre-implementation**.

## 1.1 MinVersion floor — **9.5.0**

**Decision:** Set `MinVersion` to `9.5.0` in `internal/fleet/agentlesspolicy/models.go` (constant) and align `capabilities.go` comments to the same floor.

**Rationale:**

- `/api/fleet/managed_integrations` was verified on a **9.5.0-SNAPSHOT** Kibana build (see `design.md` Decision 8).
- This matches `policyshape.MinVersionCondition` (`9.5.0`), so a separate `SupportsCondition` runtime gate becomes redundant (removed in task 4.2).
- Using `9.3.0` (the old `agentless_policies` floor) would allow plans against stacks that have the deprecated surface but not `managed_integrations`, producing 404s.

**Code touchpoints (task 1.1):**

- `internal/fleet/agentlesspolicy/models.go` — `MinVersion`, `GetVersionRequirements` error text/comments.
- `internal/fleet/agentlesspolicy/capabilities.go` — comment alignment only (no separate constant; task 4.2 removes the condition gate).

## 1.2 `KibanaHTTPAPIsManagedIntegration` ↔ schema mapping

Reviewed `generated/kbapi/kibana.gen.go` (`KibanaHTTPAPIsManagedIntegration`, lines ~50017–50091) against `internal/fleet/agentlesspolicy/schema.go` and `models.go`.

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
| `inputs` | `inputs` | Map keyed by input type ID (same as create/mapped wire) |
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

| Topic | API | Current schema | Follow-up |
|-------|-----|----------------|-----------|
| `global_data_tags` | `[{name, value: string\|number}]` | `ListNestedAttribute{name, value:string}` | Task 5.4 → `MapNestedAttribute` with `string_value` / `number_value` |
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
- Sending a full-replace PUT when the diff is confined to these flags would either be a no-op at best or risk unintended side effects; the spec requires **no API call** for such changes.
- The existing comparison (all API-backed fields equal, excluding provider plumbing and computed timestamps) remains correct once `name` and `package.version` become updatable — both are already in the comparison chain.
- Full-replace does **not** simplify this to a body-equality check: the plan never includes server-only timestamps, and building the PUT body just to compare would still require a Read.

**Implementation note for task 7:** Keep `onlyCreateOnlyFlagsChanged`; drop the GET+echo-current prelude only when the short-circuit does not fire.
