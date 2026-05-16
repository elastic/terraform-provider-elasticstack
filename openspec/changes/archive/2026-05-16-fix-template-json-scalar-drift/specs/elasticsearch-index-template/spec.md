## MODIFIED Requirements

### Requirement: Index settings semantic equality (REQ-039)

`template.settings` SHALL be modeled with a custom Plugin Framework string type that implements `basetypes.StringValuableWithSemanticEquals`. The semantic equality comparator SHALL parse both sides as JSON, flatten nested objects to dotted keys, prefix any unprefixed keys with `index.`, stringify all values, and compare the resulting maps. Two `settings` strings SHALL be considered equal whenever they represent the same effective set of index settings, regardless of dotted-vs-nested key form or the presence of an `index.` prefix on individual keys.

When Elasticsearch returns a stringified scalar echo for a setting value, the comparator SHALL treat that value as semantically equal to the practitioner-authored scalar JSON value when the effective setting is otherwise unchanged. This equivalence SHALL include JSON `null`, so a practitioner-authored `null` setting value SHALL compare equal to an Elasticsearch `"null"` string echo.

#### Scenario: Dotted vs nested keys equivalent

- GIVEN configured settings `{"index": {"number_of_shards": 1}}` and a refreshed value `{"index.number_of_shards": "1"}`
- WHEN plan runs after refresh
- THEN no diff SHALL be reported for `template.settings`

#### Scenario: `index.` prefix normalization

- GIVEN configured settings `{"refresh_interval": "1s"}` and a refreshed value `{"index.refresh_interval": "1s"}`
- WHEN plan runs after refresh
- THEN no diff SHALL be reported for `template.settings`

#### Scenario: Null scalar echo is semantically equal

- GIVEN configured settings include a JSON `null` scalar value
- AND a refreshed value returns the same effective setting as the string scalar `"null"`
- WHEN plan runs after refresh
- THEN no diff SHALL be reported for `template.settings`

#### Scenario: Invalid JSON object rejected

- GIVEN `template.settings` configured with a non-object JSON literal (e.g. an array)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error diagnostic

### Requirement: JSON and object mapping (REQ-018–REQ-025)

`metadata` SHALL be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. `template.mappings` and `template.settings` SHALL be validated as JSON objects by schema and parsed into objects during create/update. `template.alias.filter` SHALL be validated as JSON by schema and parsed into an object when non-empty during create/update. `template.alias` SHALL be mapped as a set keyed by alias name in API payload/state conversion. Set membership SHALL be determined by the alias element's semantic equality (see REQ-031); two alias values that differ only in API-derived `index_routing` or `search_routing` SHALL be treated as the same set member. Alias routing and flag fields SHALL be copied directly between Terraform values and API model fields. `template.lifecycle` SHALL be mapped as at most one lifecycle object with `data_retention`. `data_stream.hidden` SHALL be sent when present. `data_stream.allow_custom_routing` SHALL be sent only when `true`, except that on updates it SHALL also be sent when prior state had `allow_custom_routing=true` (8.x workaround behavior).

For `template.mappings`, the shared custom type SHALL treat Elasticsearch stringified scalar echoes as semantically equal to practitioner-authored scalar JSON values when the effective mapping value is otherwise unchanged. This equivalence SHALL apply to scalar leaf values such as booleans and numbers and SHALL suppress drift caused only by Elasticsearch returning a string form of the same scalar.

#### Scenario: Invalid metadata JSON

- GIVEN invalid `metadata` JSON
- WHEN create/update runs
- THEN the provider SHALL error before calling Put

#### Scenario: Routing-only alias remains a single set member

- GIVEN an alias configured with only `routing = "x"` (no `index_routing` or `search_routing` in config)
- WHEN refresh populates `index_routing = "x"` and `search_routing = "x"` from the API
- THEN the alias set in state SHALL contain exactly one element for that `name`

#### Scenario: Mappings boolean scalar echo is non-drifting

- GIVEN `template.mappings` is configured with a scalar boolean value
- AND Elasticsearch returns the same value as a JSON string scalar during refresh
- WHEN Terraform refreshes and plans the unchanged configuration
- THEN the provider SHALL treat the mapping values as semantically equal
- AND no diff SHALL be reported for that difference alone
