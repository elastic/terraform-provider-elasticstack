# kibana-slo-time-window-duration Specification

## Purpose
TBD - created by archiving change kibana-slo-time-window-duration-validation. Update Purpose after archive.
## Requirements
### Requirement: `OneOfWhenDependentPathExpressionEquals` validator is available for reuse
The `internal/utils/validators` package SHALL expose `OneOfWhenDependentPathExpressionEquals(dependentPathExpression path.Expression, dependentValue string, allowedValues []string) Condition`.

The validator SHALL:
- Be a no-op when the dependent attribute (resolved via `dependentPathExpression`) does not equal `dependentValue`.
- When the dependent attribute equals `dependentValue`, produce an `"Invalid Attribute Value Match"` diagnostic if the current attribute's value is not contained in `allowedValues`.
- Follow the same `Condition` struct pattern as existing constructors (`RequiredIfDependentPathExpressionOneOf`, `ForbiddenIfDependentPathExpressionOneOf`, `AllowedIfDependentPathExpressionOneOf`).

#### Scenario: Condition not triggered
- **GIVEN** `OneOfWhenDependentPathExpressionEquals(sibling "type", "rolling", ["7d","30d","90d"])`
- **WHEN** the dependent field `type` has value `"calendarAligned"`
- **THEN** the validator returns no diagnostics regardless of the current attribute's value

#### Scenario: Condition triggered with valid value
- **GIVEN** `OneOfWhenDependentPathExpressionEquals(sibling "type", "rolling", ["7d","30d","90d"])`
- **WHEN** the dependent field `type` has value `"rolling"` and the current attribute is `"30d"`
- **THEN** the validator returns no diagnostics

#### Scenario: Condition triggered with invalid value
- **GIVEN** `OneOfWhenDependentPathExpressionEquals(sibling "type", "rolling", ["7d","30d","90d"])`
- **WHEN** the dependent field `type` has value `"rolling"` and the current attribute is `"4d"`
- **THEN** the validator returns an `"Invalid Attribute Value Match"` diagnostic naming `[7d, 30d, 90d]` and the actual value `"4d"`

