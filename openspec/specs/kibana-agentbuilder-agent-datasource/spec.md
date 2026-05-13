# `elasticstack_kibana_agentbuilder_agent` (data source) — Schema and Functional Requirements

Data source implementation: `internal/kibana/agentbuilderagent`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_agent` data source, which exports an existing Agent Builder agent from Kibana. The data source can return the agent with or without its tool dependencies, depending on the `include_dependencies` setting.

## Schema

```hcl
data "elasticstack_kibana_agentbuilder_agent" "example" {
  agent_id = <required, string>

  space_id             = <optional, string>  # defaults to "default"
  include_dependencies = <optional, bool>    # defaults to false

  # Computed
  id             = <computed, string>         # composite: "<space_id>/<agent_id>"
  name           = <computed, string>
  description    = <computed, string>
  avatar_color   = <computed, string>
  avatar_symbol  = <computed, string>
  labels         = <computed, set(string)>
  instructions   = <computed, string>
  tools          = <computed, list(object)>   # see Tool attributes below
}
```

### Tool attributes (computed, nested)

```hcl
object {
  id                        = <computed, string>   # composite: "<space_id>/<tool_id>"
  space_id                  = <computed, string>
  tool_id                   = <computed, string>
  type                      = <computed, string>   # esql, index_search, workflow, mcp
  description               = <computed, string>
  tags                      = <computed, set(string)>
  readonly                  = <computed, bool>
  configuration             = <computed, string>   # JSON-encoded
  workflow_id               = <computed, string>   # only for workflow-type tools; requires 9.4.0+
  workflow_configuration_yaml = <computed, string> # only for workflow-type tools; requires 9.4.0+
}
```

## Requirements

### Requirement: Read API (REQ-001)

The data source SHALL call `GET /api/agentbuilder/agents/{agentId}` to read the agent from Kibana. The call SHALL be scoped to the configured Kibana space.

#### Scenario: Read an agent

- GIVEN an `agent_id` is provided
- WHEN the data source is evaluated
- THEN the provider SHALL call the Agent Builder Agents API to fetch the agent

### Requirement: Agent not found (REQ-002)

When the API returns HTTP 404, the data source SHALL return an error diagnostic instead of silently returning empty values.

#### Scenario: Agent does not exist

- GIVEN the requested agent does not exist in Kibana
- WHEN the data source is evaluated
- THEN the provider SHALL return an error indicating the agent was not found

### Requirement: API error surfacing (REQ-003)

When the API returns any non-success status other than 404, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on read

- GIVEN the Kibana API returns an error
- WHEN the data source is evaluated
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Composite id (REQ-004)

The data source SHALL expose a computed `id` in the format `<space_id>/<agent_id>`.

#### Scenario: Computed id after read

- GIVEN a successful read
- WHEN state is populated
- THEN `id` SHALL equal `<space_id>/<agent_id>`

### Requirement: space_id default (REQ-005)

When `space_id` is not configured, the data source SHALL default to `"default"`.

#### Scenario: Omitted space_id

- GIVEN `space_id` is not set in configuration
- WHEN the data source is evaluated
- THEN the API call SHALL use the "default" space

### Requirement: Agent id parsing (REQ-006)

When `agent_id` is given as a composite id in the format `<space_id>/<agent_id>`, the data source SHALL parse the space and agent id from it. If `space_id` is also configured explicitly, the explicit value SHALL take precedence.

#### Scenario: Composite agent_id

- GIVEN `agent_id = "my-space/my-agent"` and `space_id` is not set
- WHEN the data source is evaluated
- THEN the provider SHALL fetch the agent from the "my-space" space

#### Scenario: Composite agent_id with explicit space_id

- GIVEN `agent_id = "my-space/my-agent"` and `space_id = "other-space"`
- WHEN the data source is evaluated
- THEN the provider SHALL fetch the agent from the "other-space" space

### Requirement: include_dependencies default (REQ-007)

When `include_dependencies` is not configured, the data source SHALL default to `false`.

#### Scenario: Omitted include_dependencies

- GIVEN `include_dependencies` is not set
- WHEN the data source is evaluated
- THEN tool rows SHALL contain only `id`, `space_id`, and `tool_id`

### Requirement: Tool rows without dependencies (REQ-008)

When `include_dependencies` is `false`, the data source SHALL populate `tools` with minimal tool rows containing only `id` (composite), `space_id`, and `tool_id`. All other tool attributes SHALL be null.

#### Scenario: Basic tool export

- GIVEN `include_dependencies = false`
- WHEN the data source is evaluated
- THEN each tool SHALL have only `id`, `space_id`, and `tool_id` populated

### Requirement: Full tool export with dependencies (REQ-009)

When `include_dependencies` is `true`, the data source SHALL call the Agent Builder Tools API for each tool referenced by the agent, populate all tool attributes, and for workflow-type tools, fetch the referenced workflow to populate `workflow_id` and `workflow_configuration_yaml`.

#### Scenario: Full tool export

- GIVEN `include_dependencies = true`
- WHEN the data source is evaluated
- THEN each tool SHALL have all attributes populated, including configuration JSON and workflow YAML for workflow-type tools

### Requirement: Avatar field mapping (REQ-010)

When reading from the API, the data source SHALL preserve empty strings for `avatar_color` and `avatar_symbol` (not convert them to null). When the API returns null for these fields, the data source SHALL store null.

#### Scenario: Avatar field round-trip

- GIVEN the API returns `avatar_color = ""` and `avatar_symbol = null`
- WHEN the data source populates state
- THEN `avatar_color` SHALL be `""` and `avatar_symbol` SHALL be null

### Requirement: Version compatibility (REQ-011)

The data source SHALL verify the Kibana server version is at least 9.3.0 before performing the API call. If the server version is below 9.3.0, the data source SHALL fail with an error.

#### Scenario: Older Kibana version

- GIVEN the Kibana server is below version 9.3.0
- WHEN the data source is evaluated
- THEN the provider SHALL return a version error diagnostic

### Requirement: Workflow dependency version check (REQ-012)

When `include_dependencies` is `true` and the agent has workflow-type tools, the data source SHALL verify the Kibana server version is at least 9.4.0-SNAPSHOT. If the server is below that version, the data source SHALL fail with an "Unsupported server version" error because the workflow API is required to export workflow YAML.

#### Scenario: Workflow tools on older server

- GIVEN `include_dependencies = true` and the agent has workflow-type tools
- WHEN the server version is below 9.4.0-SNAPSHOT
- THEN the provider SHALL return an "Unsupported server version" error

### Requirement: Data source-level Kibana connection override (REQ-013)

The data source SHALL support a `kibana_connection` block that overrides the provider's default Kibana client for that data source instance. When the block is present, the data source SHALL use the configured connection for all API calls.

#### Scenario: Data source-level connection

- GIVEN a `kibana_connection` block is configured on the data source
- WHEN the data source is evaluated
- THEN the provider SHALL use the connection defined in the block instead of the provider default
