## MODIFIED Requirements

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)

On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, and
`metadata` back into normalized JSON strings and store them in state. For `actions`, the resource
SHALL marshal the API response into a normalized JSON string, but when the API response contains
the redacted string sentinel at a nested path and the prior known Terraform `actions` JSON value
(state on refresh, or plan on read-after-write after create or update) has a concrete value of any
JSON type at the same path, the resource SHALL preserve that prior concrete value at that path in
the final stored JSON. When no prior concrete value exists for a redacted `actions` path, or when
the prior value at that path is itself the redacted string sentinel, the resource SHALL store the
API value as returned. When the API response includes a non-nil `transform`, the resource SHALL
marshal it to a normalized JSON string and store it in state. When the API response has a nil
`transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the
remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`)
directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API
response. JSON fields SHALL normalize semantically equivalent JSON so formatting-only changes do
not create perpetual diffs.

When the Elasticsearch API response contains an empty `metadata` field, and
the prior known Terraform `metadata` value (plan on read-after-write, or state on refresh)
is the JSON string `"null"`, the resource SHALL preserve `"null"` in the `metadata` state
attribute. When the API response contains an empty `metadata` field and the prior known
value is anything other than `"null"`, the resource SHALL store the empty-object string `"{}"`
in the `metadata` state attribute. This preserves round-trip consistency for configurations that
set `metadata = jsonencode(null)`.

#### Scenario: empty metadata returned by API is stored as JSON null when prior value is null

- **GIVEN** the configuration sets `metadata = jsonencode(null)` (the string `"null"`)
- **AND** Elasticsearch receives null metadata and returns an empty metadata field in the Get Watch
  response
- **WHEN** read runs after create or during a subsequent refresh
- **THEN** the `metadata` attribute in Terraform state SHALL be `"null"`
- **AND** a subsequent plan SHALL be empty (no perpetual diff)

#### Scenario: empty-object metadata returned by API is stored as JSON empty object

- **GIVEN** the configuration sets `metadata` to the empty JSON object string `"{}"` (or omits
  `metadata` so the schema default `"{}"` is used)
- **AND** Elasticsearch returns an empty `metadata` field in the Get Watch response
- **WHEN** read runs
- **THEN** the `metadata` attribute in Terraform state SHALL be `"{}"`
