# Implementation notes — fleet-managed-integration

OpenSpec change `fleet-managed-integration`: **all tasks (1–12) complete**, plus **aggregated review fixes** (2026-07-22). Change is not archived.

## Review fixes (2026-07-22)

| Area | Change |
|------|--------|
| `cloud_connector` read | `applyCloudConnectorFromAPI` merges GET `enabled` / `cloud_connector_id`; preserves write-only `name` / `target_csp` from prior; builds block from API on import |
| Secret refs | `secrets_reconcile.go` preserves prior/plan plaintext for top-level, input, and stream vars when GET returns `{id,isSecretRef}` (bare or wrapped) |
| `policy_template` | Schema/spec: create-only, preserved on refresh, null on import; import scenario no longer claims “all attributes” from GET |
| Plan modifiers | `schema_test.go` wiring + `plan_modifier_behavior_test.go` (non-null update State/Plan); live PlanOnly acc + `ConfigPlanChecks.PreApply` incompatible with terraform-plugin-testing for non-`cloud_connector` RequiresReplace attrs |
| Create envelope | `create_test.go` + `TestNewKibanaResource_Create_readAfterWriteByDefault` (Kibana envelope read-after-write on Create) |
| Update coverage | `update_vars` adds dual `additional_datastreams_permissions`, stream input vars, and a second `global_data_tags` entry; `vars_json` / `var_group_selections` stay `deployment = "aws"` (unchanged from create) — no live var-group switch |

### Review-fix validation

| Command | Result |
|---------|--------|
| `make lint` / `make check-lint` / `make build` | exit 0 |
| `go test ./internal/fleet/managedintegration/... ./internal/clients/fleet/... -count=1` (no `TF_ACC`) | pass |
| `go test ./internal/entitycore/... -run 'KibanaResource\|SkipReadAfterWrite' -count=1` | 103 pass |
| `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-managed-integration --type change` | valid |
| `source .env && TF_ACC=1 go test ./internal/fleet/managedintegration/... -count=1 -timeout 30m` | **164 pass, 11 skip**, 0 fail |

## Task 12 — CHANGELOG and validation (complete)

### 12.1 CHANGELOG

Edited the existing `## [Unreleased]` entry introduced by [#4034](https://github.com/elastic/terraform-provider-elasticstack/pull/4034) in place (no second entry): `elasticstack_fleet_agentless_policy` → `elasticstack_fleet_managed_integration`, Kibana floor **9.5.0+**, wording “managed integrations”. Released sections untouched.

Recommended PR `## Changelog` block (see `.github/pull_request_template.md`):

```md
Customer impact: none
Summary: Rename unreleased Fleet agentless policy resource to elasticstack_fleet_managed_integration backed by managed_integrations APIs (Kibana 9.5.0+).
```

### 12.2–12.5 Validation (2026-07-22)

| Step | Command | Result |
|------|---------|--------|
| Lint | `make lint` | exit 0 |
| Build | `make build` | exit 0 |
| Full lint | `make check-lint` | exit 0 (after CHANGELOG commit; `check-fmt` requires clean tree after `fmt`) |
| OpenSpec | `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-managed-integration --type change` | `Change 'fleet-managed-integration' is valid` |
| Unit | `go test ./internal/fleet/managedintegration/... ./internal/clients/fleet/... -count=1` | pass (2 packages) |
| Acceptance | `source .env && TF_ACC=1 go test ./internal/fleet/managedintegration/... -count=1 -timeout 30m` | **152 pass, 11 skip**, 0 fail |

Acceptance skips (live-stack / cloud preconditions, Kibana ≥ 9.5.0 matrix): `TestAccResourceManagedIntegration`, `TestAccResourceManagedIntegration_CloudConnector*`, `TestAccResourceManagedIntegration_ConditionRoundTrip`, `TestAccResourceManagedIntegration_ForceDelete`, `TestAccResourceManagedIntegration_GlobalDataTagsNumber`, `TestAccResourceManagedIntegration_InputsUpdateInPlace`, `TestAccResourceManagedIntegration_NameUpdateInPlace`, `TestAccResourceManagedIntegration_NonDefaultSpace`, `TestAccResourceManagedIntegration_PackageVersionUpdate`. Unit and plan-only acceptance paths ran green on the current `.env` stack.

### Unit vs cloud coverage (final review)

| Behavior | Unit / offline | Cloud acceptance |
|----------|----------------|------------------|
| Secret ref reconciliation (multi-id, wrapped, prior list) | `secrets_reconcile_test.go` | Cloud connector acc: Terraform plaintext in state + `testCheckManagedIntegrationExternalIDStoredAsSecretRefOnAPI` (GET union decode) |
| RequiresReplace on `policy_id`, `namespace`, `policy_template`, `package.name`, `space_ids`, `cloud_connector` object | `schema_test.go` + `plan_modifier_behavior_test.go` | `CloudConnectorRequiresReplace`: DestroyBeforeCreate plan, new `cloud_connector.name`, same connector ID, new `policy_id` |
| `var_group_selections` | kbapi round-trip / convert tests | Acc keeps `deployment = "aws"` on create and update; `testCheckManagedIntegrationUpdateExtrasPersisted` still requires `var_group_selections.deployment=aws` on GET even when unchanged |
| Kibana envelope read-after-write on Create | `TestNewKibanaResource_Create_readAfterWriteByDefault` | N/A (entitycore) |

---

## Task 1 (pre-implementation)

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

Acceptance fixtures use `elasticstack_fleet_managed_integration`; tasks **10–11** renamed tests/fixtures and added live-stack coverage (see **Task 11** below).

## ~~Temporary schema vs update-body mismatch~~ (closed in task 7)

Task 5.1/5.2 made `name` and `package.version` updatable in Terraform; task 7.3 includes both in every full-replace PUT body from plan. Live rename/version bump sign-off is covered by task **11.3–11.4** acceptance tests when Kibana ≥ 9.5.0 and cloud preconditions pass.

## Task 11 — acceptance suite (complete)

Tasks **11.1–11.8** are implemented under `internal/fleet/managedintegration/acc_test.go` with helpers in `acc_kibana_version_test.go`, `acc_package_helpers_test.go`, `acc_api_assertions_test.go`, and related `*_test.go` files.

**Live matrix notes:**

- Positive `TestAccResourceManagedIntegration*` cases **skip** when Kibana is below `managedintegration.MinVersion` (`9.5.0-SNAPSHOT`), using the same `KibanaScopedClient.EnforceMinVersion` path as production.
- **`TestAccResourceManagedIntegration_VersionSkipGating`** (negative version acceptance) requires a stack **older than 9.5.0**; it passes on such stacks and is **skipped** when Kibana already meets the floor. Outcome is therefore **old-Kibana / manual matrix / CI stack-version dependent**, not a universal green on every acceptance environment.
- Cloud-hosted topology scenarios still require `skipUnlessConfirmedCloud` and Kibana ≥ 9.5.0 plus a pinned `cloud_security_posture` package (`skipUnlessManagedIntegrationLiveStack`).

## 1.1 MinVersion floor — **9.5.0-SNAPSHOT** (user-facing **9.5.0**)

**Decision:** Set `MinVersion` to `9.5.0-SNAPSHOT` in `internal/fleet/managedintegration/models.go`. Task 4.2 removed the separate condition capability gate; there is no `capabilities.go` in this package anymore.

**Rationale:**

- `/api/fleet/managed_integrations` was verified on a **9.5.0-SNAPSHOT** Kibana build (see `design.md` Decision 8).
- This shares the 9.5.0 core with `policyshape.MinVersionCondition`, so a separate `SupportsCondition` runtime gate is redundant (removed in task 4.2).
- Using `9.3.0` (the old `agentless_policies` floor) would allow plans against stacks that have the deprecated surface but not `managed_integrations`, producing 404s.

**Code touchpoints (task 1.1 / task 4):**

- `internal/fleet/managedintegration/models.go` — `MinVersion`, `GetVersionRequirements` error text/comments; `TestMinVersion_matchesPolicyshapeMinVersionConditionCore` guards core alignment with `policyshape.MinVersionCondition`.

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
| `inputs` | `inputs` | Map with caller/API-provided keys (Fleet's `"<policy_template>-<input_type>"` convention, or bare `input_type` when the policy template is empty); keys are passed through opaquely |
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
