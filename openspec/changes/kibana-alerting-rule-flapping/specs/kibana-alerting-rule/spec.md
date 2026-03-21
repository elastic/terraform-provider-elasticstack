## ADDED Requirements

### Requirement: Compatibility — `flapping` (REQ-036)

When the stack version is strictly below **8.16.0**, if the `flapping` block is present with known attribute values, create and update SHALL fail with a diagnostic that states `flapping` is only supported for Kibana **8.16** or higher (or equivalent wording naming the minimum version the provider enforces).

#### Scenario: Flapping on old stack

- GIVEN server version &lt; 8.16.0 and a configured `flapping` block with known values
- WHEN create or update runs
- THEN the provider SHALL return a flapping unsupported error

### Requirement: Validation — `flapping` block (REQ-037)

When the practitioner configures a `flapping` block, **`look_back_window`** and **`status_change_threshold`** SHALL both be set to known **integer** values (Terraform **int64**). Configuring **`enabled`** alone without both integers SHALL be invalid. The provider SHALL enforce mutual requirement of the two integer attributes when the block is present (for example via `AlsoRequires` or equivalent).

#### Scenario: Missing threshold with block present

- GIVEN a `flapping` block with `look_back_window` set and `status_change_threshold` unset (or the symmetric case)
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic

#### Scenario: Only enabled set

- GIVEN a `flapping` block with only `enabled` set
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Write path — `flapping` on create and update (REQ-038)

When the practitioner configures `flapping` with valid values, create and update requests to Kibana SHALL include a `flapping` object whose JSON properties correspond to the configured attributes (`look_back_window`, `status_change_threshold`, and `enabled` when set).

When the practitioner does **not** configure `flapping`, the provider SHALL **omit** the `flapping` property from the **update** request body so Kibana does not alter existing rule-level flapping state for that field.

#### Scenario: Create sends flapping

- GIVEN a valid `flapping` configuration and stack version ≥ 8.16.0
- WHEN create runs
- THEN the create request body SHALL include `flapping` with the configured values

#### Scenario: Update omits flapping when unset in config

- GIVEN an update where `flapping` is not configured in Terraform
- WHEN update runs
- THEN the update request body SHALL NOT include a `flapping` property

### Requirement: Acceptance tests — `flapping` (REQ-039)

The acceptance test suite for `elasticstack_kibana_alerting_rule` SHALL include at least one test case that configures `flapping`, skipped when the target stack version is strictly below **8.16.0**, that asserts create and update succeed and that Terraform state matches the configured `look_back_window` and `status_change_threshold` (and `enabled` when used).

#### Scenario: Version-gated acceptance coverage

- GIVEN a stack at **8.16.0** or newer
- WHEN the flapping acceptance test runs
- THEN it SHALL apply a configuration with `flapping` and check the expected attribute values in state after apply
