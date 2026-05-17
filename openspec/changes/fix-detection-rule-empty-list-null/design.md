## Context

The `elasticstack_kibana_security_detection_rule` resource fails with "Provider produced inconsistent result after apply" when practitioners set any of the seven nested list attributes to `[]`. The Kibana API cannot distinguish between an absent field and an empty array — it silently omits these fields from responses when they are empty. The provider's read path (`r.read()`) calls `initializeAllFieldsToDefaults()` which sets each list to `types.ListNull(...)`, then the `update*FromAPI` helpers return `types.ListNull(...)` when the API omits or returns an empty slice. This conflicts with the Terraform Plugin Framework requirement that `Optional`-only list attributes return their planned value unchanged when the planned value was a known empty list.

**Root cause chain:**

1. Practitioner writes `actions = []` → Terraform plans `cty.ListValEmpty(...)`.
2. `Create()` / `Update()` calls `r.read()` which calls `data.initializeAllFieldsToDefaults()`, setting every affected attr to `types.ListNull(elemType)`.
3. `updateActionsFromAPI()` (and the six analogous helpers) return `types.ListNull(...)` when the API carries zero items.
4. Terraform compares `null` (returned by provider) against the plan's `[]` and raises the inconsistency error.
5. Identical failure in `Read()`: prior state records `null`, config has `[]`, causing perpetual diff on every plan.

**Reference — correct pattern already in codebase:**

`updateThreatFiltersFromAPI` already handles nil-vs-empty correctly: nil pointer → `types.ListNull(...)`, pointer to zero-length slice → `types.ListValueMust(...)`. That distinction is available because the Go type is a pointer. For the seven affected attributes the Go types are non-pointer slices / structs, so `len == 0` cannot distinguish "field absent" from "field present but empty". Plan/prior-state reconciliation is therefore the correct approach.

## Goals

- Fix "Provider produced inconsistent result after apply" for all seven top-level list attributes when set to `[]`.
- Fix the same issue for `threat[*].technique` and `threat[*].technique[*].subtechnique` within a non-empty `threat` block.
- Preserve backward compatibility: attributes set to `null` (or absent) in config continue to have `null` in state.
- No schema changes required.

## Non-Goals

- Adding `Computed: true` to affected schema attributes (Approach 2 from research — evaluated and rejected; see Decisions).
- Fixing `required_fields` and `investigation_fields` (also use `types.ListNull` for empty, but are not in the reported attribute list and are out of scope per the research comment).
- Changing the Kibana API client layer.
- Adding semantic-equality custom types.

## Decisions

| Topic | Decision |
|---|---|
| Fix approach | Plan/prior-state reconciliation (Approach 1). Introduce `reconcileEmptyListsFromPlan(reference, target *Data)` in `models.go`. When `reference.X` is a known, non-null empty list and `target.X` is null, copy `reference.X` into `target.X`. |
| Caller — Create | After `readData, diags := r.read(...)`, call `reconcileEmptyListsFromPlan(&data, readData)` where `data` holds the plan. |
| Caller — Read | After `readData, diags := r.read(...)`, call `reconcileEmptyListsFromPlan(&data, readData)` where `data` holds the prior state from `req.State`. |
| Caller — Update | After `readData, diags := r.read(...)`, call `reconcileEmptyListsFromPlan(&data, readData)` where `data` holds the plan from `req.Plan`. |
| Nested `threat[*].technique` and `threat[*].technique[*].subtechnique` | Fix `convertThreatToModel` in `models_from_api_type_utils.go` to return `types.ListValueMust(getThreatTechniqueElementType(), []attr.Value{})` and `types.ListValueMust(getThreatSubtechniqueElementType(), []attr.Value{})` respectively when the API returns nil or zero-length slices inside an existing threat entry. This is safe because `technique` and `subtechnique` only appear within a concrete threat entry, so always returning empty list (not null) is consistent. The outer `Threat` attribute is handled by `reconcileEmptyListsFromPlan`. |
| Schema change | No schema change. `Computed: true` (Approach 2) is explicitly rejected: it relaxes the schema contract and hides potential misconfigurations. |
| Affected attributes (reconciliation) | `Actions`, `ExceptionsList`, `SeverityMapping`, `RiskScoreMapping`, `RelatedIntegrations`, `Threat`, `ThreatMapping`. |

## Non-Goals (implementation)

- Do not add `Computed: true` to any of the seven affected schema attributes.
- Do not reconcile `RequiredFields`, `investigation_fields`, or other list attributes not listed in the issue.

## Risks / Trade-offs

- **Read with stale prior state**: If a practitioner previously forced `actions = null` in state and later changes config to `actions = []` without triggering an update (e.g., during a standalone `terraform refresh`), the first Read after the config change will preserve `null` from prior state (the reconciliation reference is prior state, which is `null`). This is resolved on the next `terraform apply` (which runs Create or Update, where the reference is the plan). This edge case is acceptable and documented.
- **ThreatMapping**: `threat_mapping` has no else-branch in `updateFromThreatMatchRule` (it only sets the field when non-empty). The reconciliation call covers this, copying the plan's empty list when the API returns zero items and the plan had `[]`.

## Open Questions

1. Does `Update()` use `r.read()` to rebuild state after the update API call? (Confirmed from source: yes, `update.go` calls `r.read()` for the authoritative post-update read.)
2. For `threat[*].technique[*].subtechnique` — if the practitioner explicitly sets `technique = []` on a threat entry that already has techniques in state, is there a perpetual diff beyond the null/empty issue? Should be covered by an acceptance test.
3. Are there other rule-type-specific list attributes (beyond the 7 reported) that exhibit the same null/empty bug? (Candidates: `required_fields`, `investigation_fields` — noted as out of scope per research.)
