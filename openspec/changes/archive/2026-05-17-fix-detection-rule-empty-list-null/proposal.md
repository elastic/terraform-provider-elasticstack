## Why

The `elasticstack_kibana_security_detection_rule` resource produces **"Provider produced inconsistent result after apply"** errors when a practitioner explicitly sets any of the following nested list attributes to `[]` ([#2006](https://github.com/elastic/terraform-provider-elasticstack/issues/2006)):

| Attribute | Affected location |
|---|---|
| `actions` | `models_from_api_type_utils.go` — `updateActionsFromAPI` |
| `exceptions_list` | `models_from_api_type_utils.go` — `updateExceptionsListFromAPI` |
| `severity_mapping` | `models_from_api_type_utils.go` — `updateSeverityMappingFromAPI` |
| `risk_score_mapping` | `models_from_api_type_utils.go` — `updateRiskScoreMappingFromAPI` |
| `related_integrations` | `models_from_api_type_utils.go` — `updateRelatedIntegrationsFromAPI` |
| `threat` | `models_from_api_type_utils.go` — `updateThreatFromAPI` |
| `threat_mapping` | `models_threat_match.go` — `updateFromThreatMatchRule` |
| `threat[*].technique[*].subtechnique` | `models_from_api_type_utils.go` — `convertThreatToModel` |

The root cause is a null-vs-empty mismatch: Terraform plans the value as `cty.ListValEmpty(...)` but the provider's read-back (via `r.read()` → `updateCommonRuleFieldsFromAPI()`) returns `types.ListNull(...)` when the Kibana API omits or returns an empty array for those fields. The Terraform Plugin Framework invariant for `Optional`-only list attributes requires the provider to return exactly the planned value when the planned value was a known empty list.

## What Changes

- Add a `reconcileEmptyListsFromPlan` helper to `models.go` that copies explicit empty lists from a reference `Data` (plan or prior state) into the post-read `Data` for each of the seven top-level affected attributes.
- Call `reconcileEmptyListsFromPlan` in `Create()`, `Read()`, and `Update()` after the `r.read()` call.
- Extend the reconciliation logic for nested `threat[*].technique` and `threat[*].technique[*].subtechnique` so explicitly configured empty lists remain `[]` in state while omitted / `null` values remain `null`.
- Add or extend acceptance test coverage to verify that `terraform apply` with all seven attributes set to `[]` succeeds and produces a consistent empty-list state.

No schema changes (no `Computed: true` additions). This is a behaviour-only bug fix.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-security-detection-rule`: Add null-to-empty-list reconciliation for the seven nested list attributes, plus nested `threat` technique/subtechnique reconciliation that preserves null-vs-empty semantics, so `terraform apply` with `attribute = []` does not produce a "Provider produced inconsistent result after apply" error (REQ-033–REQ-036).

## Impact

- **Specs**: Delta under `openspec/changes/fix-detection-rule-empty-list-null/specs/kibana-security-detection-rule/spec.md` until merged into canonical spec.
- **Implementation** (future): `internal/kibana/security_detection_rule/models.go` (new/extended reconciliation helper), `create.go`, `read.go`, `update.go` (one helper call each), `models_from_api_type_utils.go` (nested threat mapping support), `acc_test.go` (new acceptance test step).
