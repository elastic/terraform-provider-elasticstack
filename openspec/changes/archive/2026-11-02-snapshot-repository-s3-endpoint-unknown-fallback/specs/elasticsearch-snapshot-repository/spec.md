## MODIFIED Requirements

### Requirement: S3 endpoint plan drift prevention (REQ-017)

The implementer SHALL determine whether the Elasticsearch GET `/_snapshot/{name}` response
returns `endpoint` and `path_style_access` in the S3 settings object.

- If the GET response **does** return a field, no schema change is required and state will
  round-trip correctly for that attribute.
- If the GET response **does not** return a field (i.e. Elasticsearch treats it as a
  write-only or omitted setting), the `settingsToS3` read function SHALL implement read-side
  state inheritance: when the API omits the field, the prior state value SHALL be preserved
  in the refreshed state rather than overwriting it with `null` (for `endpoint`) or the schema
  default (for `path_style_access`). This mirrors the existing `compressFallback` pattern
  used in `settingsToFs` and `settingsToURL`.

Read-side state inheritance SHALL apply only when the prior value (the prior state value, or
the planned value during the read that follows create or update) is known. During the read
that follows create or update, the prior value is the planned value, and an Optional +
Computed attribute that is not configured is unknown in the plan. When the API omits a field
and the prior value is unknown, the refreshed state SHALL hold `null` for `endpoint` and the
schema default (`false`) for `path_style_access`. State SHALL never hold an unknown value for
either attribute after create, update, read, or import.

The decision and its rationale SHALL be documented in the implementing PR description.

#### Scenario: No spurious diff after second apply

- GIVEN a repository created with `endpoint` set
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes for the `endpoint` attribute

#### Scenario: No spurious diff for path_style_access after second apply

- GIVEN a repository created with `path_style_access = true`
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes for the `path_style_access` attribute

#### Scenario: Create without endpoint

- GIVEN an `s3` repository block that does not set `endpoint`
- WHEN the resource is created
- THEN the apply SHALL succeed
- AND state SHALL hold `null` for `s3.endpoint`
- BUT state SHALL NOT hold an unknown value for `s3.endpoint`

#### Scenario: Update without endpoint

- GIVEN an existing repository whose `s3` block does not set `endpoint`
- WHEN another `s3` setting, such as `chunk_size`, is changed and applied
- THEN the apply SHALL succeed
- AND state SHALL hold `null` for `s3.endpoint`

#### Scenario: No spurious diff without endpoint

- GIVEN a repository created from an `s3` block that does not set `endpoint`
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes

#### Scenario: Endpoint removed together with another change

- GIVEN an existing repository whose `s3` block sets `endpoint`
- WHEN `endpoint` is removed from the `s3` block and another `s3` setting is changed in the same apply
- THEN the apply SHALL succeed
- AND state SHALL hold `null` for `s3.endpoint`

#### Scenario: API value wins over an unknown prior value

- GIVEN the prior `endpoint` value is unknown
- AND the GET response returns `endpoint`
- WHEN the resource reads the repository
- THEN state SHALL hold the `endpoint` value from the GET response

#### Scenario: Known prior endpoint is still inherited

- GIVEN the prior `endpoint` value is a known URL
- AND the GET response does not return `endpoint`
- WHEN the resource reads the repository
- THEN state SHALL hold the prior `endpoint` URL
