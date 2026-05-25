## ADDED Requirements

### Requirement: global_data_tags entry must have exactly one value (REQ-GDT-001)

Each entry in the `global_data_tags` map attribute of `elasticstack_fleet_agent_policy` MUST
have exactly one of `string_value` or `number_value` set to a non-null value. An entry with
neither `string_value` nor `number_value` set is not valid.

The schema MUST enforce this constraint at **plan time** using `AtLeastOneOf` validators on
both `string_value` and `number_value`, referencing the sibling attribute paths
(`string_value` and `number_value`) within the same nested object. Both attributes MUST
retain their existing `ConflictsWith` validators preventing *both* from being set
simultaneously.

The effect is that exactly one of the two attributes must be set for each `global_data_tags`
entry: neither-set and both-set are both validation errors.

#### Scenario: Neither value field set is rejected at plan time

- GIVEN a Terraform configuration for `elasticstack_fleet_agent_policy`
- AND `global_data_tags` contains an entry with no `string_value` and no `number_value`
  (e.g., `"my_tag" = {}`)
- WHEN `terraform plan` runs
- THEN Terraform MUST emit a validation diagnostic error indicating that at least one of
  `string_value` or `number_value` must be set
- AND `terraform apply` MUST NOT be reached for that configuration

#### Scenario: Only string_value set is valid

- GIVEN a `global_data_tags` entry with `string_value = "foo"` and no `number_value`
- WHEN `terraform plan` runs
- THEN validation MUST succeed for that entry

#### Scenario: Only number_value set is valid

- GIVEN a `global_data_tags` entry with `number_value = 42` and no `string_value`
- WHEN `terraform plan` runs
- THEN validation MUST succeed for that entry

#### Scenario: Both values set is rejected at plan time

- GIVEN a `global_data_tags` entry with both `string_value = "foo"` and `number_value = 42`
- WHEN `terraform plan` runs
- THEN Terraform MUST emit a validation diagnostic error (existing `ConflictsWith` behavior,
  unchanged)

---

### Requirement: global_data_tags conversion must not panic on null-null entries (REQ-GDT-002)

The internal `convertGlobalDataTags` function MUST NOT dereference a nil pointer when
converting a `global_data_tags` map entry to the Kibana Fleet API model.

If a `global_data_tags` entry has both `string_value` and `number_value` null or unknown at
apply time (a state-corruption or API-inconsistency edge case that bypasses plan-time
validators), the function MUST emit a `diag.Diagnostics` error with summary
`"Invalid global_data_tags entry"` and return without panicking.

The conversion logic MUST use explicit `IsNull()` / `IsUnknown()` checks on both
`item.StringValue` and `item.NumberValue` rather than relying on pointer-nil tests of
`ValueStringPointer()` / `ValueFloat32Pointer()` to determine which API value type to use.

#### Scenario: Null-null entry returns error diagnostic, not panic

- GIVEN an `agentPolicyModel` with a `global_data_tags` map entry where `string_value` is
  null and `number_value` is null
- WHEN `convertGlobalDataTags` is called
- THEN the function MUST return a `diag.Diagnostics` with at least one error
- AND the function MUST NOT panic
- AND the returned diagnostics MUST include a summary of `"Invalid global_data_tags entry"`
