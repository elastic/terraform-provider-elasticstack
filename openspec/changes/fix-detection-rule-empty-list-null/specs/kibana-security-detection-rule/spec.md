## MODIFIED Requirements

### Requirement: Empty-list consistency for optional nested list attributes (REQ-100)

When a practitioner explicitly configures any of the following `elasticstack_kibana_security_detection_rule` attributes as an empty list (`[]`), the provider SHALL return an empty list — not `null` — for that attribute in state after `Create`, `Read`, and `Update`. This preserves the Terraform Plugin Framework invariant for `Optional`-only list attributes: the provider MUST return the planned value unchanged when the planned value is a known, non-null empty list.

Affected attributes:

| Attribute | Schema type |
|---|---|
| `actions` | `ListNestedAttribute` |
| `exceptions_list` | `ListNestedAttribute` |
| `severity_mapping` | `ListNestedAttribute` |
| `risk_score_mapping` | `ListNestedAttribute` |
| `related_integrations` | `ListNestedAttribute` |
| `threat` | `ListNestedAttribute` |
| `threat_mapping` | `ListNestedAttribute` |

#### Scenario: Apply with all affected attributes set to empty list

- GIVEN a resource configuration with `actions = []`, `exceptions_list = []`, `severity_mapping = []`, `risk_score_mapping = []`, `related_integrations = []`, `threat = []`, and `threat_mapping = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL succeed without a "Provider produced inconsistent result after apply" diagnostic
- AND each of the seven attributes SHALL be stored as an empty list (`[]`) in Terraform state

#### Scenario: Subsequent plan shows no diff for empty-list attributes

- GIVEN a successfully applied resource with any of the seven attributes stored as `[]` in state
- WHEN `terraform plan` runs without any configuration change
- THEN the plan SHALL be empty (no changes) for those attributes

#### Scenario: Null configuration is preserved

- GIVEN a resource configuration where one or more of the seven attributes is absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` (not `[]`) for those attributes in state

### Requirement: Empty-list consistency for nested `threat` sub-lists (REQ-101)

When a `threat` block is configured with one or more entries, and a practitioner explicitly configures `technique = []` for a threat entry, the provider SHALL return an empty list (`[]`) — not `null` — for `technique` in state. The same rule applies when a practitioner explicitly configures `subtechnique = []` within a technique entry.

If `technique` or `subtechnique` is absent from configuration or explicitly `null`, the provider SHALL preserve `null` for that attribute in state and SHALL NOT normalize it to `[]`.

#### Scenario: Threat entry with explicitly empty techniques preserves empty list

- GIVEN a resource configuration with one `threat` entry and `technique = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `[]` for `technique` in state for that threat entry
- AND the provider SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Threat entry with omitted or null techniques preserves null

- GIVEN a resource configuration with one `threat` entry and `technique` absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` for `technique` in state for that threat entry

#### Scenario: Technique entry with explicitly empty subtechniques preserves empty list

- GIVEN a resource configuration with a threat entry containing a technique entry and `subtechnique = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `[]` for `subtechnique` in state for that technique entry
- AND the provider SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Technique entry with omitted or null subtechniques preserves null

- GIVEN a resource configuration with a threat entry containing a technique entry and `subtechnique` absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` for `subtechnique` in state for that technique entry

### Requirement: Reconciliation helper for plan/state alignment (REQ-102)

The provider SHALL implement a `reconcileEmptyListsFromPlan` function (or equivalent logic) in the `securitydetectionrule` package. For each of the seven affected attributes, this function SHALL: if the reference value (plan for Create/Update, prior state for Read) is a known, non-null empty list AND the post-read value is null, replace the post-read null with the reference empty list.

This function SHALL be called after each `r.read()` invocation in `Create`, `Read`, and `Update`.

#### Scenario: Null in post-read is overwritten when reference has empty list

- GIVEN a reference `Data` where `Actions` is a known empty list and `target.Actions` is null
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL be set to an empty list identical to `reference.Actions`

#### Scenario: Non-null target is not overwritten

- GIVEN a reference `Data` where `Actions` is a known empty list and `target.Actions` is a non-empty list with items
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL remain unchanged

#### Scenario: Null reference does not overwrite null target

- GIVEN a reference `Data` where `Actions` is null and `target.Actions` is null
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL remain null

### Requirement: Acceptance test — empty-list round-trip (REQ-103)

The acceptance test suite SHALL include a test that exercises the empty-list scenario for all seven affected attributes in a single resource configuration. The test SHALL apply a configuration with all seven attributes set to `[]`, assert that `terraform apply` succeeds without "inconsistent result" diagnostics, assert that all seven attributes are stored as empty lists in state, and assert that a subsequent `terraform plan` produces an empty plan.

#### Scenario: Acceptance test apply with empty lists succeeds

- GIVEN a resource configuration with `actions = []`, `exceptions_list = []`, `severity_mapping = []`, `risk_score_mapping = []`, `related_integrations = []`, `threat = []`, and `threat_mapping = []`
- WHEN the acceptance test runs `terraform apply`
- THEN `terraform apply` SHALL succeed without any "Provider produced inconsistent result after apply" diagnostics
- AND the test SHALL verify that each of the seven attributes is stored as an empty list in state

#### Scenario: No-op plan after empty-list apply

- GIVEN a successfully applied rule with the seven attributes stored as empty lists
- WHEN the acceptance test runs a second `terraform plan`
- THEN the plan SHALL be empty (no proposed changes)
