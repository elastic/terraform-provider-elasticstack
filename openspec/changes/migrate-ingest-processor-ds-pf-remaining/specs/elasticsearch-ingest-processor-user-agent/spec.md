## MODIFIED Requirements

### Requirement: Common processor fields
The `elasticstack_elasticsearch_ingest_processor_user_agent` data source SHALL expose `description`, `if`, `ignore_failure`, `on_failure`, and `tag` as optional attributes. When configured, each SHALL be included in the serialized JSON. When not configured, each SHALL be omitted from the JSON (except `ignore_failure`, which defaults to `false` and is always included).

#### Scenario: Common fields in schema
- GIVEN the data source schema definition
- WHEN inspecting available attributes
- THEN `description`, `if`, `ignore_failure`, `on_failure`, and `tag` SHALL be valid optional attributes

#### Scenario: Common fields included in JSON when configured
- GIVEN a configuration that sets `description = "parse user agent"`, `if = "ctx.agent != null"`, `ignore_failure = true`, `on_failure = ['{"set":{"field":"error.message","value":"ua failed"}}']`, and `tag = "ua-tag"`
- WHEN the data source is read
- THEN `json` SHALL include `"description": "parse user agent"`, `"if": "ctx.agent != null"`, `"ignore_failure": true`, `"on_failure": [{"set":{"field":"error.message","value":"ua failed"}}]`, and `"tag": "ua-tag"`

#### Scenario: Common fields omitted when not configured
- GIVEN a configuration that does not set `description`, `if`, `on_failure`, or `tag`
- WHEN the data source is read
- THEN `json` SHALL omit `"description"`, `"if"`, `"on_failure"`, and `"tag"` keys
- AND `json` SHALL include `"ignore_failure": false`

## ADDED Requirements

### Requirement: description field
The data source SHALL accept an optional `description` string attribute. When configured, it SHALL be included in the serialized JSON under the `"description"` key. When not configured, the key SHALL be omitted.

#### Scenario: description configured
- GIVEN `description = "Parse user agent string"`
- WHEN the data source is read
- THEN `json` SHALL include `"description": "Parse user agent string"`

### Requirement: if field
The data source SHALL accept an optional `if` string attribute. When configured, it SHALL be included in the serialized JSON under the `"if"` key. When not configured, the key SHALL be omitted.

#### Scenario: if configured
- GIVEN `if = "ctx.agent != null"`
- WHEN the data source is read
- THEN `json` SHALL include `"if": "ctx.agent != null"`

### Requirement: on_failure field
The data source SHALL accept an optional `on_failure` list of JSON string attributes. When configured with one or more elements, each element SHALL be parsed as JSON and included in the serialized JSON under the `"on_failure"` key as an array of objects. When not configured, the key SHALL be omitted.

#### Scenario: on_failure configured
- GIVEN `on_failure = ['{"set":{"field":"error.message","value":"user agent failed"}}']`
- WHEN the data source is read
- THEN `json` SHALL include `"on_failure": [{"set":{"field":"error.message","value":"user agent failed"}}]`

### Requirement: tag field
The data source SHALL accept an optional `tag` string attribute. When configured, it SHALL be included in the serialized JSON under the `"tag"` key. When not configured, the key SHALL be omitted.

#### Scenario: tag configured
- GIVEN `tag = "user-agent-parse"`
- WHEN the data source is read
- THEN `json` SHALL include `"tag": "user-agent-parse"`
