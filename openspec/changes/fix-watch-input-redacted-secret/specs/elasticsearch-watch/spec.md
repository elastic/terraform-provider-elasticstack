## MODIFIED Requirements

### Requirement: Input redaction preservation (REQ-030)

On read, the resource SHALL preserve prior known Terraform values of any JSON type at nested `input`
paths where the Watcher API returns the redacted string sentinel (`::es_redacted::`), mirroring the
existing `actions` redaction-preservation behavior (REQ-014–016, REQ-023–027). The prior `input`
JSON SHALL be the last-applied value from Terraform state when read is a refresh, and the configured
value from the Terraform plan when read runs as read-after-write after create or update. When no
prior concrete value exists for a redacted `input` path, or when the prior value at that path is
itself the redacted string sentinel, the resource SHALL store the API value as returned.

The `fromAPIModel` function SHALL accept the prior `input` value and, when that prior value is
known, unmarshal both the API `input` response and the prior Terraform `input` value, call
`mergePreserveRedactedLeaves` on the two `map[string]any` trees, and store the merged result in
state. All non-redacted `input` fields SHALL remain authoritative from the API response.

#### Scenario: Redacted HTTP input basic auth password on create read-after-write

- **GIVEN** the Terraform plan for `apply` includes a concrete sensitive `password` value at
  `input.http.request.auth.basic.password`
- **WHEN** the provider performs read-after-write after a successful Put Watch call and the Get Watch
  API returns `::es_redacted::` at `input.http.request.auth.basic.password`
- **THEN** the resource SHALL store the prior plan concrete `password` value at that path in state
- **AND** all other `input` fields from the API response SHALL be stored in state unchanged
- **AND** a subsequent `terraform plan` SHALL produce an empty diff

#### Scenario: Redacted HTTP input basic auth password on refresh

- **GIVEN** the last-applied `input` JSON in Terraform state includes a concrete sensitive `password`
  value at `input.http.request.auth.basic.password`
- **WHEN** read runs as a refresh and the Get Watch API returns `::es_redacted::` at that path
- **THEN** the resource SHALL preserve the prior state concrete `password` value at that path
- **AND** non-redacted `input` fields from the API response SHALL still be reflected in state

#### Scenario: No prior input value on import or first read

- **GIVEN** Terraform has no prior concrete `input` value in state or plan for a redacted nested path
  (e.g. after `terraform import`)
- **WHEN** read runs and the Get Watch API returns `::es_redacted::` at a nested `input` path
- **THEN** the resource SHALL store the redacted API value at that path in state

#### Scenario: Redacted input path when prior is a non-string

- **GIVEN** the prior `input` JSON has a non-string value at a nested path (e.g. an object or array)
- **WHEN** read runs and the Get Watch API returns `::es_redacted::` at that same path
- **THEN** the resource SHALL preserve the prior non-string value at that path
- **AND** non-redacted `input` fields from the API response SHALL still be reflected in state

### MODIFIED Requirement: JSON field mapping — read/state (REQ-023–027)

The narrative SHALL be updated to include `input` alongside `actions` in the redaction-preservation
description. Specifically, the sentence:

> For `actions`, the resource SHALL marshal the API response into a normalized JSON string, but when
> the API response contains the redacted string sentinel at a nested path and the prior known
> Terraform `actions` JSON value ... has a concrete value of any JSON type at the same path, the
> resource SHALL preserve that prior concrete value at that path in the final stored JSON.

SHALL be extended to add:

> The same redaction-preservation behavior applies to `input`: when the API response contains the
> redacted string sentinel at a nested `input` path and the prior known Terraform `input` JSON value
> (state on refresh, or plan on read-after-write after create or update) has a concrete value of any
> JSON type at the same path, the resource SHALL preserve that prior concrete value at that path in
> the final stored `input` JSON. When no prior concrete value exists for a redacted `input` path, or
> when the prior value at that path is itself the redacted string sentinel, the resource SHALL store
> the API value as returned.
