## MODIFIED Requirements

### Requirement: Validation — filter_id (REQ-008)

The `filter_id` attribute SHALL be validated at plan time to contain at least one character, to
contain only lowercase alphanumeric characters (a–z and 0–9), hyphens, underscores, and dots, and
to start and end with an alphanumeric character. No upper-bound length restriction is applied.
The validator SHALL use `ml.IDValidatorWithoutLength()` (defined in
`internal/elasticsearch/ml/idvalidator.go`) instead of `ml.IDValidator()`.

#### Scenario: Long filter_id accepted

- GIVEN a `filter_id` that is longer than 64 characters and otherwise valid
- WHEN the configuration is validated
- THEN validation SHALL succeed with no error diagnostics

#### Scenario: Invalid filter_id — uppercase characters

- GIVEN a configuration with `filter_id = "INVALID_ID"`
- WHEN the configuration is validated
- THEN validation SHALL fail with an appropriate error diagnostic

#### Scenario: Empty filter_id rejected

- GIVEN a `filter_id` that is an empty string
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic
