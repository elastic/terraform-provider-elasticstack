## MODIFIED Requirements

### Requirement: Config validation — privilege check SHALL be deferred for unknown kibana attributes (REQ-028)

The provider SHALL skip per-element privilege validation for any `kibana` set
element where: the element object itself (`types.Object`) is unknown, OR the
`feature` attribute of the element is unknown, OR the `base` attribute of the
element is unknown.

In those cases, validation SHALL be deferred to apply time. The apply-time guard
inside `expandKibana` SHALL continue to enforce the invariant (exactly one of
`base` or `feature` must be non-empty) once all values are resolved.

For elements where all controlling attributes are known, the existing constraint
SHALL continue to be enforced at config-validation time:

- If both `base` and `feature` are non-empty, the provider SHALL return an error
  ("Only one of the `feature` or `base` privileges allowed!").
- If both `base` and `feature` are empty, the provider SHALL return an error
  ("Either one of the `feature` or `base` privileges must be set for kibana role!").

This requirement enables `dynamic` blocks on `kibana.feature` (and `kibana.base`),
where `for_each` is evaluated after config validation and the resulting attributes
are unknown during the `ValidateResourceConfig` phase.

#### Scenario: Dynamic feature block — plan succeeds

- GIVEN a `kibana` block that uses `dynamic "feature" { for_each = … }` to
  conditionally emit `feature` sub-blocks
- WHEN Terraform runs the plan phase (`ValidateResourceConfig`)
- THEN the provider SHALL NOT return a privilege-validation error

#### Scenario: Dynamic feature block — apply enforces constraint

- GIVEN a `kibana` block with a `dynamic "feature"` block whose `for_each`
  evaluates to an empty list at apply time, and no `base` is configured
- WHEN Terraform runs apply
- THEN the provider SHALL return an error requiring at least one of `base` or
  `feature` to be set

#### Scenario: Static config — missing privilege still rejected at plan time

- GIVEN a `kibana` block with neither `base` nor `feature` configured (no dynamic
  block; both are known-empty sets at config-validation time)
- WHEN Terraform runs the plan phase
- THEN the provider SHALL return an error ("Either one of the `feature` or `base`
  privileges must be set for kibana role!")

#### Scenario: Unknown element object — validation deferred

- GIVEN a `kibana` set element whose entire `types.Object` is unknown (e.g.
  `for_each` references a completely unknown collection)
- WHEN Terraform runs the plan phase
- THEN the provider SHALL NOT return a privilege-validation error for that element
