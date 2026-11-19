## ADDED Requirements

### Requirement: Plan-time `scheduled_task_id` unknown on `enabled` transition (REQ-054)

`scheduled_task_id` SHALL plan as unknown whenever the rule's effective `enabled` value in the new plan
differs from the effective `enabled` value in prior state, in either direction (disabled → enabled or
enabled → disabled). The comparison SHALL use the planned value of `enabled` after default resolution,
not the raw configuration value. When `enabled` is unchanged, `scheduled_task_id` SHALL retain its
prior state value per `UseStateForUnknown`, unless superseded by REQ-026.

#### Scenario: Enabling a never-run rule does not produce an inconsistent result

- **GIVEN** a rule in state with `enabled = false` and `scheduled_task_id` null
- **WHEN** the practitioner changes configuration to `enabled = true` and runs `terraform apply`
- **THEN** the plan SHALL show `scheduled_task_id` as unknown
- **AND** apply SHALL succeed, with the API's returned `scheduled_task_id` accepted without an
  inconsistent-result error

#### Scenario: Disabling a running rule does not produce an inconsistent result

- **GIVEN** a rule in state with `enabled = true` and a known non-null `scheduled_task_id`
- **WHEN** the practitioner changes configuration to `enabled = false` and runs `terraform apply`
- **THEN** the plan SHALL show `scheduled_task_id` as unknown
- **AND** apply SHALL succeed without an inconsistent-result error

#### Scenario: Omitting `enabled` while already enabled is a no-op

- **GIVEN** a rule in state with `enabled = true` (resolved from the default, or explicit)
- **WHEN** the practitioner removes the `enabled` line from configuration entirely (so the default
  `true` applies) and runs `terraform plan`
- **THEN** the plan SHALL show no change to `scheduled_task_id`

#### Scenario: Enabling via config omission still unknowns `scheduled_task_id`

- **GIVEN** a rule in state with `enabled = false` and `scheduled_task_id` null
- **WHEN** the practitioner removes the `enabled` line from configuration entirely, so the default
  `true` resolves in the plan, and runs `terraform apply`
- **THEN** the plan SHALL show `scheduled_task_id` as unknown
- **AND** apply SHALL succeed without an inconsistent-result error

#### Scenario: Unknown planned `enabled` conservatively unknowns `scheduled_task_id`

- **GIVEN** a rule in state with a known `enabled` value
- **WHEN** `enabled` in the new plan is unknown (e.g. derived from another resource not yet applied)
- **THEN** the plan SHALL show `scheduled_task_id` as unknown rather than erroring or treating it as
  unchanged

#### Scenario: Unrelated update with `enabled` unchanged preserves `scheduled_task_id`

- **GIVEN** a rule in state with `enabled = true` and a known non-null `scheduled_task_id`
- **WHEN** the practitioner changes an unrelated attribute (e.g. `interval`) with `enabled` unchanged
  and runs `terraform plan`
- **THEN** the plan SHALL show no change to `scheduled_task_id`
