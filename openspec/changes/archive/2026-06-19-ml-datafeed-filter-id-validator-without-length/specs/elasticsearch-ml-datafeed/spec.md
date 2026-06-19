## MODIFIED Requirements

### Requirement: datafeed_id validation (REQ-021)

The `datafeed_id` attribute SHALL be validated at plan time to contain at least one character, to
contain only lowercase alphanumeric characters (a–z and 0–9), hyphens, underscores, and dots, and
to start and end with an alphanumeric character. No upper-bound length restriction is applied.
The validator SHALL use `ml.IDValidatorWithoutLength()` (defined in
`internal/elasticsearch/ml/idvalidator.go`) instead of `ml.IDValidator()`.

#### Scenario: Long datafeed_id accepted

- GIVEN a `datafeed_id` that is longer than 64 characters and otherwise valid (e.g. `datafeed-opserv-riskviewxml-customer-transaction-volume-decline-stop` at 68 characters)
- WHEN the configuration is validated
- THEN validation SHALL succeed with no error diagnostics

#### Scenario: Invalid datafeed_id rejected — illegal characters

- GIVEN a `datafeed_id` that starts with a hyphen or contains uppercase characters
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic referencing the character-class restriction

#### Scenario: Empty datafeed_id rejected

- GIVEN a `datafeed_id` that is an empty string
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic
