## ADDED Requirements

### Requirement: Compatibility — `flapping` (REQ-036)

When the stack version is strictly below **8.16.0**, if the `flapping` block is present with known attribute values, create and update SHALL fail with a diagnostic that states `flapping` is only supported for Kibana **8.16** or higher (or equivalent wording naming the minimum version the provider enforces).

#### Scenario: Flapping on old stack

- GIVEN server version &lt; 8.16.0 and a configured `flapping` block with known values
- WHEN create or update runs
- THEN the provider SHALL return a flapping unsupported error

### Requirement: Compatibility — `flapping.enabled` (REQ-040)

When the stack version is strictly below **9.3.0**, if **`flapping.enabled`** is set to a known value in configuration, create and update SHALL fail with a diagnostic that states `flapping.enabled` is only supported for Elastic Stack **9.3** or higher (or equivalent wording naming the minimum version the provider enforces).

#### Scenario: Enabled on pre-9.3 stack

- GIVEN server version &lt; 9.3.0 and `flapping.enabled` configured with a known value (with the two integer attributes also set as required)
- WHEN create or update runs
- THEN the provider SHALL return a `flapping.enabled` unsupported error

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

When the practitioner configures `flapping` with valid values, create and update requests to Kibana SHALL include a `flapping` object whose JSON properties correspond to the configured attributes (`look_back_window`, `status_change_threshold`, and `enabled` when set and the stack satisfies REQ-040).

When the practitioner does **not** configure `flapping`, the provider SHALL **omit** the `flapping` property from the **update** request body so Kibana does not alter existing rule-level flapping state for that field.

#### Scenario: Create sends flapping

- GIVEN a valid `flapping` configuration, stack version ≥ 8.16.0, and stack version ≥ 9.3.0 when `enabled` is configured
- WHEN create runs
- THEN the create request body SHALL include `flapping` with the configured values (including `enabled` in JSON only when configured and REQ-040 is satisfied)

#### Scenario: Update omits flapping when unset in config

- GIVEN an update where `flapping` is not configured in Terraform
- WHEN update runs
- THEN the update request body SHALL NOT include a `flapping` property

### Requirement: Acceptance tests — `flapping` (REQ-039)

The acceptance test suite for `elasticstack_kibana_alerting_rule` SHALL include:

1. At least one test case that configures **`flapping`** with **`look_back_window`** and **`status_change_threshold`** only (no `enabled`), skipped when the target stack is strictly below **8.16.0**, that asserts create and update succeed and that state matches the two integers.
2. When asserting **`flapping.enabled`**, the test SHALL be skipped unless the stack is **9.3.0** or newer (for example via `SkipFunc` or an equivalent minimum version for those steps only).

If existing tests only cover `flapping` together with `enabled`, additional tests or steps SHALL be added so that (1) is satisfied on **8.16.0+** stacks without requiring **9.3.0**.

#### Scenario: Integer-only flapping on 8.16+

- GIVEN a stack at **8.16.0** or newer and below **9.3.0** (or a test run that skips `enabled` steps)
- WHEN the integer-only flapping acceptance test runs
- THEN it SHALL apply a configuration with `flapping` without `enabled` and check the expected `look_back_window` and `status_change_threshold` in state after apply

#### Scenario: Enabled gated at 9.3+

- GIVEN a stack at **9.3.0** or newer
- WHEN the flapping acceptance test that configures `flapping.enabled` runs
- THEN it SHALL assert create/update and state for `enabled` as configured

#### Scenario: Enabled steps skipped below 9.3

- GIVEN a stack strictly below **9.3.0**
- WHEN acceptance tests that configure `flapping.enabled` are evaluated
- THEN those steps SHALL be skipped (and integer-only coverage from scenario “Integer-only flapping on 8.16+” SHALL still apply where applicable)
