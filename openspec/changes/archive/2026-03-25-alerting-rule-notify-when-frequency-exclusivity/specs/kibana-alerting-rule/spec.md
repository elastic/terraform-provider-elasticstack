## ADDED Requirements

### Requirement: Plan normalization — top-level `notify_when` with action `frequency` (REQ-041)

When **any** `actions` entry includes a `frequency` block in the Terraform configuration, and the **planned** value for top-level `notify_when` is **unknown** (not yet known at plan time), the provider SHALL set the planned top-level `notify_when` to **null** so that create/update does not send rule-level `notify_when` in that case. When top-level `notify_when` is **known** in the plan (including an explicit practitioner value or a known value from state), the provider SHALL NOT use this normalization to clear or override that known value.

#### Scenario: Unknown top-level notify_when with action frequency

- GIVEN a configuration where at least one `actions` block includes `frequency` and the plan has top-level `notify_when` unknown
- WHEN the plan is finalized for the resource
- THEN the planned top-level `notify_when` SHALL be null and the provider SHALL NOT send rule-level `notify_when` on create/update solely from that unknown planned value

#### Scenario: Explicit top-level notify_when unchanged by normalization

- GIVEN a configuration where top-level `notify_when` is planned as a **known** string (from practitioner input or from state)
- WHEN the plan is finalized
- THEN REQ-041 normalization SHALL NOT replace or clear that known planned value (REQ-042 governs invalid combinations of rule-level `notify_when` or `throttle` with action `frequency`)

### Requirement: Validation — rule-level notification vs action `frequency` (REQ-042)

The provider SHALL reject configuration at plan/validate time when **both** of the following are true:

1. The practitioner configures **either** top-level `notify_when` to a **known** non-null value with a **non-empty** string, **or** top-level `throttle` to a **known** **non-empty** string (mirroring Kibana documentation: rule-level `notify_when` or `throttle` cannot be combined with per-action `frequency` parameters).

2. **Any** `actions` entry includes a `frequency` block.

Configurations that include action `frequency` but **omit** top-level `notify_when` and top-level `throttle` (or leave them null / empty as applicable) SHALL remain valid. Configurations that set top-level `notify_when` or `throttle` but **no** action `frequency` SHALL remain valid. REQ-014 (notify_when required before 8.6) and REQ-015 (frequency only from 8.6) remain in force; REQ-042 only applies when **both** sides in (1) and (2) are satisfied, so it does not impose `notify_when` on stacks below 8.6 when the practitioner is not using action `frequency`.

#### Scenario: Rule-level notify_when and action frequency

- GIVEN top-level `notify_when` set to a non-empty known value and at least one `actions` entry with a `frequency` block
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic explaining that rule-level `notify_when` (or `throttle`) cannot be combined with action `frequency`

#### Scenario: Rule-level throttle and action frequency

- GIVEN top-level `throttle` set to a non-empty known value and at least one `actions` entry with a `frequency` block
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic explaining that rule-level `notify_when` or `throttle` cannot be combined with action `frequency`

#### Scenario: Action frequency only

- GIVEN at least one `actions` entry with a `frequency` block and neither top-level `notify_when` nor top-level `throttle` set to a non-empty known value
- WHEN Terraform validates configuration
- THEN the provider SHALL NOT reject the configuration under REQ-042

#### Scenario: Rule-level notify_when without action frequency

- GIVEN top-level `notify_when` set to a non-empty known value and no `frequency` block on any action
- WHEN Terraform validates configuration
- THEN the provider SHALL NOT reject the configuration under REQ-042
