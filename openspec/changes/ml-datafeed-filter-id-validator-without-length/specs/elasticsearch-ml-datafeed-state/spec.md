## MODIFIED Requirements

### Requirement: datafeed_id validation (REQ-021)

The `datafeed_id` attribute SHALL be validated at plan time to contain at least one character and to
contain only alphanumeric characters, hyphens, and underscores. No upper-bound length restriction
is applied. The validator SHALL use `ml.IDValidatorWithoutLength()` (defined in
`internal/elasticsearch/ml/idvalidator.go`) instead of `ml.IDValidator()`.

#### Scenario: Long datafeed_id accepted

- GIVEN a `datafeed_id` that is longer than 64 characters and otherwise valid
- WHEN the configuration is validated
- THEN validation SHALL succeed with no error diagnostics

#### Scenario: Invalid datafeed_id rejected — illegal characters

- GIVEN a `datafeed_id` containing characters outside alphanumeric, hyphens, and underscores
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic

#### Scenario: Empty datafeed_id rejected

- GIVEN a `datafeed_id` that is an empty string
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic
