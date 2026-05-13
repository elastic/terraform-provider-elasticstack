## MODIFIED Requirements

### Requirement: `time_window.duration` description is accurate
The description for `time_window.duration` SHALL list the exact permitted values per window type, replacing the previous inaccurate text.

**Current (inaccurate):** *"Any duration greater than 1 day can be used: days, weeks, months, quarters, years."*

**Replacement:** The description SHALL state:
- For `type = "rolling"`: duration must be one of `7d` (7 days), `30d` (30 days), or `90d` (90 days).
- For `type = "calendarAligned"`: duration must be either `1w` (weekly) or `1M` (monthly).

#### Scenario: Rolling window description is accurate
- **WHEN** a practitioner reads the `time_window.duration` attribute description for a rolling window
- **THEN** the description names `7d`, `30d`, and `90d` as the only valid values
- **AND** the description does not claim that arbitrary day-based durations are accepted

#### Scenario: Calendar aligned window description is accurate
- **WHEN** a practitioner reads the `time_window.duration` attribute description for a calendar-aligned window
- **THEN** the description names `1w` and `1M` as the only valid values

---

### Requirement: `time_window.duration` is validated at plan time for `rolling` windows
The `duration` attribute in the `time_window` block SHALL be validated at Terraform plan time when `type = "rolling"`.

#### Scenario: Valid rolling duration passes validation
- **GIVEN** a `time_window` block with `type = "rolling"`
- **WHEN** `duration` is set to `"7d"`, `"30d"`, or `"90d"`
- **THEN** Terraform plan succeeds with no validation error on `duration`

#### Scenario: Invalid rolling duration fails at plan time
- **GIVEN** a `time_window` block with `type = "rolling"`
- **WHEN** `duration` is set to any value other than `"7d"`, `"30d"`, or `"90d"` (e.g. `"4d"`, `"14d"`, `"1w"`)
- **THEN** Terraform plan produces a diagnostic with summary `"Invalid Attribute Value Match"` targeting the `duration` attribute
- **AND** the diagnostic message names `"7d"`, `"30d"`, `"90d"` as the valid values for `type = "rolling"`
- **AND** no API call is made

---

### Requirement: `time_window.duration` is validated at plan time for `calendarAligned` windows
The `duration` attribute in the `time_window` block SHALL be validated at Terraform plan time when `type = "calendarAligned"`.

#### Scenario: Valid calendarAligned duration passes validation
- **GIVEN** a `time_window` block with `type = "calendarAligned"`
- **WHEN** `duration` is set to `"1w"` or `"1M"`
- **THEN** Terraform plan succeeds with no validation error on `duration`

#### Scenario: Invalid calendarAligned duration fails at plan time
- **GIVEN** a `time_window` block with `type = "calendarAligned"`
- **WHEN** `duration` is set to any value other than `"1w"` or `"1M"` (e.g. `"30d"`, `"1Q"`)
- **THEN** Terraform plan produces a diagnostic with summary `"Invalid Attribute Value Match"` targeting the `duration` attribute
- **AND** the diagnostic message names `"1w"`, `"1M"` as the valid values for `type = "calendarAligned"`
- **AND** no API call is made

---

## ADDED Requirements

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
