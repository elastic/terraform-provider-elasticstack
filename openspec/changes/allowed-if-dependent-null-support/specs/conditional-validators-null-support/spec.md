## ADDED Requirements

### Requirement: AllowedIf validators accept unknown dependent values
`AllowedIf` conditional validators SHALL treat an unknown dependent value as satisfying the dependency condition so validation does not fail while the dependent attribute is still computed.

#### Scenario: Dependent value is unknown
- **WHEN** an `AllowedIf` validator evaluates a configured attribute and the dependent attribute value is unknown
- **THEN** the validator SHALL return no validation error for that dependency state

### Requirement: AllowedIf equals-or-null helper
The validators package SHALL expose a helper that allows an attribute when a dependent path equals a required string value or when the dependent path is null/unset.

#### Scenario: Dependent value equals required value
- **WHEN** the equals-or-null `AllowedIf` helper evaluates a configured attribute and the dependent path equals the required value
- **THEN** the validator SHALL return no validation error

#### Scenario: Dependent value is null
- **WHEN** the equals-or-null `AllowedIf` helper evaluates a configured attribute and the dependent path is null or unset
- **THEN** the validator SHALL return no validation error

#### Scenario: Dependent value is a different concrete value
- **WHEN** the equals-or-null `AllowedIf` helper evaluates a configured attribute and the dependent path has a known value that does not equal the required value
- **THEN** the validator SHALL return a validation error
