## 1. Schema — add `agent_policy_ids`, make `agent_policy_id` Optional

- [ ] 1.1 In `internal/fleet/elastic_defend_integration_policy/schema.go`:
  - Change `agent_policy_id` from `Required: true` to `Optional: true` and add a
    `stringvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression())` validator.
  - Add a new `agent_policy_ids` `schema.ListAttribute` (Optional, `ElementType: types.StringType`)
    with validators `listvalidator.ConflictsWith(path.Root("agent_policy_id").Expression())` and
    `listvalidator.SizeAtLeast(1)`.
  - Add the required imports: `"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"`,
    `"github.com/hashicorp/terraform-plugin-framework/path"` (framework path).

## 2. Model — add `AgentPolicyIDs` field

- [ ] 2.1 In `internal/fleet/elastic_defend_integration_policy/models.go`, add
  `AgentPolicyIDs types.List \`tfsdk:"agent_policy_ids"\`` to `elasticDefendIntegrationPolicyModel`
  immediately after the `AgentPolicyID` field.

## 3. Version gate — `MinVersionPolicyIDs` constant and capability check

- [ ] 3.1 In `internal/fleet/elastic_defend_integration_policy/resource.go` (or a new
  `capabilities.go` file in the same package), declare:
  ```go
  var MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))
  ```
  Add the `"github.com/hashicorp/go-version"` import.
- [ ] 3.2 In `internal/fleet/elastic_defend_integration_policy/create.go` (`Create`), before using `agent_policy_ids` from model, check if
  `model.AgentPolicyIDs` is non-null/non-unknown and call `client.EnforceMinVersion(ctx,
  MinVersionPolicyIDs)`. Return an error diagnostic if the version gate fails.
- [ ] 3.3 Apply the same version gate in `internal/fleet/elastic_defend_integration_policy/update.go` (`Update`).

## 4. Request — populate `PolicyIds` in bootstrap and finalize requests

- [ ] 4.1 In `internal/fleet/elastic_defend_integration_policy/request.go`, update
  `buildBootstrapRequest` to:
  - When `model.AgentPolicyIDs` is non-null and non-unknown: extract the list of IDs, set
    `req.PolicyIds = &ids` and `req.PolicyId = &ids[0]` (first-element compatibility).
  - Otherwise: set `req.PolicyId = model.AgentPolicyID.ValueStringPointer()` (existing behavior).
- [ ] 4.2 Apply the same logic to `buildFinalizeRequest`.
- [ ] 4.3 Update the signatures of `buildBootstrapRequest` and `buildFinalizeRequest` to accept
  `ctx context.Context` if needed for `ElementsAs`.

## 5. Mapping — populate model from API response

- [ ] 5.1 In `internal/fleet/elastic_defend_integration_policy/mapping.go`, update
  `populateModelFromAPI` to mirror the generic resource pattern:
  - Determine which field was originally in state: `originallyUsedAgentPolicyID` (via
    `typeutils.IsKnown(model.AgentPolicyID)`) and `originallyUsedAgentPolicyIDs` (via
    `typeutils.IsKnown(model.AgentPolicyIDs)`).
  - If `originallyUsedAgentPolicyID`: set `model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)`.
  - If `originallyUsedAgentPolicyIDs`: if `policy.PolicyIds != nil`, convert and assign to
    `model.AgentPolicyIDs`; else set to `types.ListNull(types.StringType)`.
  - If neither flag is set (edge case / import): default to populating `model.AgentPolicyID`
    from `policy.PolicyId`.
  - Remove (or guard) the existing unconditional `model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)` line.

## 6. Update the canonical spec

- [ ] 6.1 In `openspec/specs/fleet-elastic-defend-integration-policy/spec.md`:
  - Update the schema block to show `agent_policy_id` as `<optional, string>` and add
    `agent_policy_ids = <optional, list(string)>` with a note about the version gate.
  - Update REQ-003 to remove the exclusion of `agent_policy_ids` and reflect that the resource
    now supports the multi-policy attribute.
  - Add a new requirement (REQ-014 or next available) for the multi-agent-policy behavior,
    version gate, and conflict semantics.
  - Apply the delta spec from
    `openspec/changes/fleet-elastic-defend-agent-policy-ids/specs/fleet-elastic-defend-integration-policy/spec.md`.

## 7. Acceptance tests

- [ ] 7.1 Add `TestAccResourceElasticDefendIntegrationPolicy_multiAgentPolicy` in
  `internal/fleet/elastic_defend_integration_policy/acc_test.go` (or the existing test file):
  - Create two agent policies.
  - Create an Elastic Defend integration policy with `agent_policy_ids = [<id1>, <id2>]`.
  - Verify both agent policies receive the Defend package policy.
  - Verify plan is clean (no diff) after apply.
  - Add an update step that changes the list and verify consistency.
- [ ] 7.2 Add a version-gate test (or skip annotation) that asserts that `agent_policy_ids`
  returns an appropriate error on stacks older than 8.15.0.
- [ ] 7.3 Verify that existing tests using `agent_policy_id` continue to pass unmodified.

## 8. Build and validation

- [ ] 8.1 `make build` — provider compiles without errors.
- [ ] 8.2 `go test ./internal/fleet/elastic_defend_integration_policy/... -v` — package tests pass.
- [ ] 8.3 `make check-openspec` — spec validation passes.
