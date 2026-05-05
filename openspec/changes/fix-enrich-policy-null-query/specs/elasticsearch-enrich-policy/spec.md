## MODIFIED Requirements

### Requirement: Query mapping (REQ-013–REQ-015)

When `query` is set in configuration, the resource SHALL send it as a parsed JSON object
in the `query` field of the Put request body. When `query` is null or not configured, the
resource SHALL omit the `query` field from the Put request body entirely. On read, when
the API response includes a `query` field that is non-null and non-empty **and whose
JSON-marshaled form is not the literal bytes `null`**, the resource SHALL store it in
state as a normalized JSON string; otherwise `query` SHALL be stored as null.

Specifically, when `json.Marshal` applied to the `*types.Query` value returned by the
go-elasticsearch typed client produces the bytes `null` (which occurs when the client
deserializes an explicit `"query": null` in the API response into a non-nil pointer to a
zero-value struct), the provider SHALL treat this identically to an absent `query` field
and SHALL store `null` in Terraform state.

#### Scenario: Query sent as JSON object

- GIVEN `query` is set to a valid JSON string
- WHEN create runs
- THEN the Put request body SHALL contain `query` as a parsed JSON object

#### Scenario: Null query omitted from request

- GIVEN `query` is not configured
- WHEN create runs
- THEN the Put request body SHALL not contain a `query` field

#### Scenario: Null query preserved in state

- GIVEN the API response has no `query` field or a null query
- WHEN read runs
- THEN `query` in state SHALL be null

#### Scenario: Marshaled-null query treated as absent

- GIVEN the API response contains `"query": null` (explicit JSON null)
- AND the go-elasticsearch typed client returns a non-nil `*types.Query` for this value
- WHEN `json.Marshal` applied to that `*types.Query` produces the bytes `null`
- THEN the provider SHALL store `query` as null in Terraform state (not the string `"null"`)

#### Scenario: No-query policy is idempotent across applies

- GIVEN an enrich policy was created with `query` omitted from configuration
- WHEN `terraform apply` is run a second time with the same configuration
- THEN Terraform SHALL plan no changes (no replacement, no update)
