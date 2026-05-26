## MODIFIED Requirements

### Requirement: Read state mapping (REQ-022–REQ-026)

On read, the resource SHALL set `name` and `version` from the API response. On read, when API `metadata` is present, it SHALL be serialized into a JSON string and stored in state. On read, when API `template` is present, it SHALL be flattened into `template` state, including aliases, mappings, and settings. User-defined alias `routing` SHALL be preserved during read/refresh, because this field may be omitted by the API response and therefore SHALL not be overwritten from response data.

For `template.mappings`, the resource SHALL treat Elasticsearch stringified scalar echoes as semantically equal to practitioner-authored scalar JSON values when the effective mapping value is otherwise unchanged. This equivalence SHALL apply to scalar leaf values such as booleans and numbers and SHALL suppress drift and post-apply consistency errors caused only by Elasticsearch returning a string form of the same scalar.

For `template.settings`, the resource SHALL treat Elasticsearch stringified scalar echoes as semantically equal to practitioner-authored scalar JSON values when the effective setting value is otherwise unchanged. This equivalence SHALL include JSON `null`, so a practitioner-authored `null` setting value SHALL compare equal to an Elasticsearch `"null"` string echo.

#### Scenario: Routing preserved on refresh

- GIVEN user-configured alias `routing` and API omits routing fields
- WHEN read runs
- THEN user `routing` SHALL not be lost from state

#### Scenario: Mappings boolean scalar echo is non-drifting

- GIVEN `template.mappings` is configured with a scalar boolean value
- AND Elasticsearch returns the same value as a JSON string scalar during refresh
- WHEN apply completes or a later refresh runs
- THEN the provider SHALL treat the mapping values as semantically equal
- AND Terraform SHALL NOT report a provider inconsistent-result error or follow-up drift for that difference alone

#### Scenario: Settings null scalar echo is non-drifting

- GIVEN `template.settings` is configured with a JSON `null` scalar value
- AND Elasticsearch returns the same value as the string scalar `"null"` during refresh
- WHEN apply completes or a later refresh runs
- THEN the provider SHALL treat the settings values as semantically equal
- AND Terraform SHALL NOT report a provider inconsistent-result error or follow-up drift for that difference alone
