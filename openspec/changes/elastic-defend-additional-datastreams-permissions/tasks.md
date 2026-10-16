## 1. Schema

- [ ] 1.1 Add `additional_datastreams_permissions` to `resourceSchema()` in `internal/fleet/elastic_defend_integration_policy/schema.go` as `schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true}` with `PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}`. Do not add a `SizeAtLeast(1)` validator — an explicit empty list must be accepted (it is the clearing mechanism, per design.md Decision 1).
- [ ] 1.2 Add a description noting that removing the attribute from configuration does not revoke previously granted permissions server-side, and that `[]` is the supported way to clear them.

## 2. Model

- [ ] 2.1 Add `AdditionalDatastreamsPermissions types.List `tfsdk:"additional_datastreams_permissions"`` to `elasticDefendIntegrationPolicyModel` in `internal/fleet/elastic_defend_integration_policy/models.go`.
- [ ] 2.2 Add `MinVersionAdditionalDatastreamsPermissions = version.Must(version.NewVersion("9.1.0"))` to `internal/fleet/elastic_defend_integration_policy/resource.go`, alongside the existing `MinVersionPolicyIDs`.
- [ ] 2.3 Restructure `elasticDefendIntegrationPolicyModel.GetVersionRequirements` (`models.go:51-54`) to accumulate requirements into a `[]entitycore.VersionRequirement` slice instead of returning `nil, nil` early when `agent_policy_ids` is unknown: append the existing `agent_policy_ids` requirement when `model.AgentPolicyIDs` is known, and append a new `additional_datastreams_permissions` requirement (`entitycore.NewAttributeVersionRequirement(path.Root("additional_datastreams_permissions"), *MinVersionAdditionalDatastreamsPermissions, ...)`) when `model.AdditionalDatastreamsPermissions` is known and non-null. Both checks must be independent so either, both, or neither requirement can apply.

## 3. Request mapping

- [ ] 3.1 In `buildBootstrapRequest` (`internal/fleet/elastic_defend_integration_policy/common.go:411`), set `req.AdditionalDatastreamsPermissions` from `model.AdditionalDatastreamsPermissions` when known (via `ElementsAs` into `*[]string`, including an empty slice); leave it `nil` when null/unknown. No revoke-on-unset branch.
- [ ] 3.2 Apply the identical mapping in `buildFinalizeRequest` (`internal/fleet/elastic_defend_integration_policy/common.go:468`), which is also the request builder used by the plain `Update` path (`update.go`), so bootstrap and finalize/update never disagree about which permissions are desired.
- [ ] 3.3 Factor the mapping from 3.1/3.2 into one small shared helper (e.g. `additionalDatastreamsPermissionsFromModel(ctx, model) (*[]string, diag.Diagnostics)`) in `common.go`, called from both builders, so the "known (incl. empty) -> send; null/unknown -> omit" rule is implemented once.

## 4. Read-back mapping

- [ ] 4.1 In `populateModelFromAPI` (`internal/fleet/elastic_defend_integration_policy/common.go:45`), map `policy.AdditionalDatastreamsPermissions`: non-empty API list -> `types.ListValueFrom(ctx, types.StringType, *data.AdditionalDatastreamsPermissions)`. When the API value is absent or empty, preserve a known empty list already present in `model.AdditionalDatastreamsPermissions` (i.e. the plan/prior state already held `[]`) as `[]`; otherwise set `types.ListNull(types.StringType)`. An import (null model) therefore yields null for an absent or empty API value.

## 5. Unit tests

- [ ] 5.1 Add request-body test cases for `buildBootstrapRequest` and `buildFinalizeRequest`: populated list, explicit empty list, and null/unknown (field omitted from the request body in all three builders/paths).
- [ ] 5.2 Add `populateModelFromAPI` round-trip tests: non-empty API list -> list state; absent/empty API list with a known `[]` in the input model -> `[]` in state; absent/empty API list with a null/unknown input model -> null in state.
- [ ] 5.3 Add `GetVersionRequirements` test cases: `additional_datastreams_permissions` configured alone, `agent_policy_ids` configured alone, both configured, and neither configured — asserting each applicable requirement is present and the other is excluded, confirming the task-2.3 restructuring doesn't lose either requirement.

## 6. Acceptance tests

- [ ] 6.1 Add `TestAccResourceElasticDefendIntegrationPolicyAdditionalDatastreamsPermissions` to `internal/fleet/elastic_defend_integration_policy/acc_test.go`, mirroring `TestAccResourceIntegrationPolicyAdditionalDatastreamsPermissions` (`internal/fleet/integration_policy/acc_test.go:269-345`):
  - Create with one permission (e.g. `logs-nginx-*`); assert `additional_datastreams_permissions.#` = `"1"` in Terraform state and a matching server-side value via `fleetclient.GetDefendPackagePolicy`.
  - Import + `ImportStateVerify`.
  - Update to two permissions; re-assert state and server-side value changed.
  - **Remove the attribute from config** (no change to any other attribute) and assert the plan retains the prior value and the server-side value is **retained**, not cleared — this is the behavior that differs from `integration_policy`. Do not assert that the request omits the field (see design.md Decision 1, plan-modifier interaction).
  - Set `additional_datastreams_permissions = []` and assert it is cleared server-side and is `[]` (not null) in state.
- [ ] 6.3 Add an out-of-band permissions test: set `additional_datastreams_permissions` on the package policy directly through the Fleet API, then apply a configuration that never sets the attribute. Assert an empty plan, state populated from the API, and the server-side value unchanged.
- [ ] 6.2 Add a best-effort version-gate acceptance test parallel to `TestAccResourceElasticDefendIntegrationPolicy_agentPolicyIDsVersionGate` (`acc_test.go:388-415`), asserting `ExpectError` when `additional_datastreams_permissions` is set against a server below 9.1.0, and that omitting the attribute still creates/updates successfully against such a server. This only runs on stacks between the Defend minimum (8.14.0) and 9.1.0 and may be skipped in CI; the task 5.3 unit tests are the primary guard.

## 7. Docs (follow-up, not required for this proposal)

- [ ] 7.1 Add `additional_datastreams_permissions` to `docs/resources/fleet_elastic_defend_integration_policy.md` and the resource's example, documenting the no-revoke-on-unset behavior and that `[]` clears previously granted permissions.

## 8. Spec sync

- [ ] 8.1 After implementation, update the attribute listing in the schema block of the main spec (and REQ-011's read-mapping text if it enumerates modeled fields), then sync the delta spec in `openspec/changes/elastic-defend-additional-datastreams-permissions/specs/fleet-elastic-defend-integration-policy/spec.md` into `openspec/specs/fleet-elastic-defend-integration-policy/spec.md` (openspec-sync-specs skill) and archive the change (openspec-archive-change skill).
