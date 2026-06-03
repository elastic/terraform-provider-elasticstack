## MODIFIED Requirements

Delta spec for capability `elasticstack-fleet-agent-policy`
(canonical spec: `openspec/specs/elasticstack-fleet-agent-policy/spec.md`).

### Requirement: space_ids must retain its configured value when the Fleet API omits the field (REQ-SID-001)

The `space_ids` attribute of `elasticstack_fleet_agent_policy` MUST retain its
configured, non-null value in state after every successful `terraform apply` — including the
initial Create and any subsequent Update — even when the Fleet API response body omits
`space_ids` (i.e., the field is absent and `kbapi.AgentPolicy.SpaceIds` is unmarshaled as `nil`).

The `populateFromAPI` function MUST NOT overwrite a non-null, non-unknown `SpaceIDs` model
value with `types.SetNull` when the API returns `nil` for `SpaceIds`.

The fix MUST apply to all callers of `populateFromAPI`: Create, Read, and Update.

#### Scenario: Create with space_ids — API omits field in response

- GIVEN an `elasticstack_fleet_agent_policy` resource configured with
  `space_ids = ["my_space"]`
- WHEN `terraform apply` runs and the Fleet POST/GET response bodies omit `space_ids`
  (nil / omitempty)
- THEN Terraform MUST NOT raise "Provider produced inconsistent result after apply"
- AND the state for `space_ids` MUST equal `["my_space"]` after apply
- AND a subsequent `terraform plan` MUST show no changes

#### Scenario: Read with space_ids — API omits field in response

- GIVEN a previously applied `elasticstack_fleet_agent_policy` with
  `space_ids = ["my_space"]` in state
- WHEN Terraform performs a Read (refresh) and the Fleet GET response omits `space_ids`
- THEN the state for `space_ids` MUST remain `["my_space"]`
- AND a subsequent `terraform plan` MUST show no changes (no spurious diff)

#### Scenario: Update with space_ids — API omits field in response

- GIVEN a previously applied `elasticstack_fleet_agent_policy` with
  `space_ids = ["my_space"]` in state
- WHEN `terraform apply` runs an in-place update for another attribute and the Fleet PUT/GET
  response bodies omit `space_ids`
- THEN the state for `space_ids` MUST remain `["my_space"]`
- AND a subsequent `terraform plan` MUST show no changes

#### Scenario: API omits space_ids and model value is null — remains null

- GIVEN an `elasticstack_fleet_agent_policy` resource configured WITHOUT `space_ids`
  (i.e., `space_ids` is unknown in the plan because the attribute is Optional+Computed)
- WHEN `populateFromAPI` is called and the API returns `nil` for `SpaceIds`
- THEN the model `SpaceIDs` MUST be set to `types.SetNull` (unchanged behaviour)
- AND no inconsistency error MUST occur
