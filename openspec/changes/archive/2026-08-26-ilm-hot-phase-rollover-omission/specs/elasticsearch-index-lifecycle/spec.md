## ADDED Requirements

### Requirement: Hot phase rollover omission preserved on read (REQ-036)

When Elasticsearch returns an empty `rollover` action (`"rollover": {}`, with no conditions) for the `hot` phase of an ILM policy, and neither the Terraform configuration nor the prior Terraform state declared a `rollover` block for that phase, the provider SHALL treat the action as absent and leave `hot.rollover` as `null` in state, rather than materializing a non-null object with all-null attributes.

When the prior Terraform state already contains a non-null `rollover` object for the `hot` phase (the user has explicitly declared a `rollover` block, including an empty one with no conditions), the provider SHALL preserve that declaration in state by continuing to write the returned action, consistent with the prior-state preservation pattern used for the `readonly`, `freeze`, and `unfollow` toggle actions (REQ-029).

This normalization applies only to the `rollover` action within the `hot` phase; it does not change how `warm`, `cold`, or `delete` phases handle their actions, and it does not change the write path — the provider already omits `rollover` from the Elasticsearch ILM PUT payload when the Terraform configuration does not declare it.

#### Scenario: Hot phase without rollover produces no diff after refresh

- GIVEN a Terraform configuration for `elasticstack_elasticsearch_index_lifecycle` whose `hot` phase declares no `rollover` block
- AND the prior Terraform state has `hot.rollover = null`
- WHEN the provider reads the policy and Elasticsearch's response includes `"rollover": {}` for the `hot` phase
- THEN the provider SHALL store `hot.rollover = null` in state
- AND a subsequent `terraform plan` SHALL show no changes

#### Scenario: Explicitly declared empty rollover is preserved

- GIVEN prior Terraform state that already declares `hot.rollover` as a non-null object (the user configured a `rollover {}` block, with or without conditions)
- WHEN the provider reads the policy and Elasticsearch's response includes `"rollover": {}` for the `hot` phase
- THEN the provider SHALL continue to write a non-null `rollover` object into state for that phase

#### Scenario: Rollover with conditions is unaffected

- GIVEN Elasticsearch returns a `hot` phase `rollover` action containing one or more trigger conditions (for example `max_age`)
- WHEN the provider flattens the phase
- THEN state SHALL contain those conditions under `hot.rollover`, unchanged by this normalization
