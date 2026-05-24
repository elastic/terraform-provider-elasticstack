# conditional-validators-null-support Specification

## Purpose
TBD - created by archiving change allowed-if-dependent-null-support. Update Purpose after archive.
## Requirements
### Requirement: AllowedIf validators accept unknown dependent values
`AllowedIf` conditional validators SHALL treat an unknown dependent value as satisfying the dependency condition so validation does not fail while the dependent attribute is still computed.

#### Scenario: Dependent value is unknown
- **WHEN** an `AllowedIf` validator evaluates a configured attribute and the dependent attribute value is unknown
- **THEN** the validator SHALL return no validation error for that dependency state

### Requirement: AllowedIf validators support explicit null handling options
The validators package SHALL require an options argument on `AllowedIf` validator constructors that explicitly controls whether a null or unset dependent path is accepted.

#### Scenario: Dependent value equals required value
- **WHEN** an options-enabled `AllowedIf` validator evaluates a configured attribute and the dependent path equals the required value
- **THEN** the validator SHALL return no validation error

#### Scenario: Dependent value is null and null is allowed
- **WHEN** an options-enabled `AllowedIf` validator evaluates a configured attribute, the dependent path is null or unset, and the options allow null
- **THEN** the validator SHALL return no validation error

#### Scenario: Dependent value is null and null is not allowed
- **WHEN** an options-enabled `AllowedIf` validator evaluates a configured attribute, the dependent path is null or unset, and the options do not allow null
- **THEN** the validator SHALL return a validation error

#### Scenario: Dependent value is a different concrete value
- **WHEN** an options-enabled `AllowedIf` validator evaluates a configured attribute and the dependent path has a known value that does not equal the required value
- **THEN** the validator SHALL return a validation error

