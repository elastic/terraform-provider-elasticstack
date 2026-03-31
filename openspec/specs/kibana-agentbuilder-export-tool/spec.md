# `elasticstack_kibana_agentbuilder_export_tool` — Schema and Functional Requirements

Data source implementation: `internal/kibana/exportagentbuilder/tool`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_export_tool` data source, which reads an Agent Builder tool by ID from Kibana and optionally exports the workflow configuration referenced by a `workflow`-type tool. Tool export requires Kibana 9.3.0 or above. Workflow export (via `include_workflow`) requires Kibana 9.4.0 or above.

## Schema

```hcl
data "elasticstack_kibana_agentbuilder_export_tool" "example" {
  id               = <required, string>         # tool ID to export; accepts composite "<space_id>/<tool_id>"
  space_id         = <optional, string>         # Kibana space; defaults to "default"
  include_workflow = <optional, bool>           # when true, also export referenced workflow (requires 9.4.0+; only valid for type = "workflow")

  # Computed
  tool_id                    = <computed, string>
  type                       = <computed, string>        # esql, index_search, workflow, or mcp
  description                = <computed, string>
  tags                       = <computed, list(string)>
  readonly                   = <computed, bool>
  configuration              = <computed, string>        # JSON-encoded tool configuration
  workflow_id                = <computed, string>        # only populated when include_workflow = true
  workflow_configuration_yaml = <computed, string>       # normalized YAML; only populated when include_workflow = true
}
```

## Requirements

### Requirement: Read API (REQ-001)

The data source SHALL use the Kibana Agent Builder Tools API to read a tool by ID. The API call SHALL be scoped to the configured Kibana space via the space-aware path request editor.

#### Scenario: Successful read

- GIVEN a tool exists in Kibana
- WHEN the data source is read
- THEN the provider SHALL call `GET /api/agentbuilder/tools/{toolId}` scoped to the configured space and populate all attributes in state

### Requirement: Tool not found (REQ-002)

When the API returns HTTP 404 for the tool, the data source SHALL return an error diagnostic indicating the tool was not found rather than silently producing empty state.

#### Scenario: Tool does not exist

- GIVEN the tool ID does not exist in Kibana
- WHEN the data source reads
- THEN the provider SHALL return a "Tool not found" error diagnostic

### Requirement: API error surfacing (REQ-003)

When the API returns a non-success status (other than 404), the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on read

- GIVEN the Kibana API returns a non-404 error
- WHEN the data source reads
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-004)

The data source SHALL set a computed `id` in the format `<space_id>/<tool_id>` after a successful read.

#### Scenario: Computed id after read

- GIVEN a successful read
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<tool_id>`

### Requirement: Composite input id (REQ-005)

When the `id` input is in the composite format `<space_id>/<tool_id>`, the data source SHALL parse it to extract the space and tool IDs. The extracted space SHALL be used for the API call unless `space_id` is also explicitly configured, in which case the explicit `space_id` SHALL take precedence.

#### Scenario: Composite id as input

- GIVEN `id` is set to `"myspace/my-tool"`
- WHEN the data source reads
- THEN the provider SHALL use `myspace` as the space and `my-tool` as the tool ID

### Requirement: space_id default (REQ-006)

When `space_id` is not configured and the input `id` is not a composite id, the data source SHALL use `"default"` as the space.

#### Scenario: Plain tool ID with no space_id

- GIVEN `id` is a plain string (not composite) and `space_id` is not set
- WHEN the data source reads
- THEN the provider SHALL use `"default"` as the space

### Requirement: Configuration JSON mapping (REQ-007)

The data source SHALL marshal the tool's `configuration` map from the API response to a JSON-encoded string for the `configuration` attribute in state.

#### Scenario: Configuration populated

- GIVEN a tool with a non-empty configuration
- WHEN the data source reads
- THEN `configuration` in state SHALL be a valid JSON string representing the tool configuration

### Requirement: Optional field mapping (REQ-008)

When the API response contains a null `description`, the data source SHALL set `description` to null in state. When the API response contains an empty `tags` list, the data source SHALL set `tags` to null in state.

#### Scenario: Absent optional fields

- GIVEN the tool has no description or tags
- WHEN the data source reads
- THEN `description` and `tags` SHALL be null in state

### Requirement: Workflow export (REQ-009)

When `include_workflow` is true, the data source SHALL fetch the workflow referenced by the tool's `workflow_id` configuration key using `GET /api/workflows/{workflowId}` scoped to the same space, and SHALL populate `workflow_id` and `workflow_configuration_yaml` in state from the response.

#### Scenario: Successful workflow export

- GIVEN a workflow-type tool with `include_workflow = true`
- WHEN the data source reads
- THEN the provider SHALL fetch the referenced workflow and populate `workflow_id` and `workflow_configuration_yaml` in state

### Requirement: include_workflow type validation (REQ-010)

When `include_workflow` is true but the tool type is not `workflow`, the data source SHALL return an error diagnostic and SHALL NOT attempt to fetch any workflow.

#### Scenario: include_workflow on non-workflow tool

- GIVEN `include_workflow = true` and the tool type is `esql`
- WHEN the data source reads
- THEN the provider SHALL return an error indicating `include_workflow` is only valid for `workflow`-type tools

### Requirement: include_workflow configuration key (REQ-011)

When `include_workflow` is true, the data source SHALL extract the `workflow_id` value from the tool's `configuration` map. If the key is missing or not a non-empty string, the data source SHALL return an error diagnostic.

#### Scenario: Missing workflow_id in configuration

- GIVEN a workflow-type tool whose configuration does not contain a `workflow_id` key
- WHEN the data source reads with `include_workflow = true`
- THEN the provider SHALL return an error diagnostic

### Requirement: Workflow not found (REQ-012)

When the referenced workflow is not found (API returns 404), the data source SHALL return a "Workflow not found" error diagnostic.

#### Scenario: Referenced workflow deleted

- GIVEN the workflow referenced by the tool no longer exists
- WHEN the data source reads with `include_workflow = true`
- THEN the provider SHALL return a "Workflow not found" error diagnostic

### Requirement: Workflow fields null when not exported (REQ-013)

When `include_workflow` is false or not set, the data source SHALL set `workflow_id` and `workflow_configuration_yaml` to null in state.

#### Scenario: include_workflow omitted

- GIVEN `include_workflow` is not set
- WHEN the data source reads
- THEN `workflow_id` and `workflow_configuration_yaml` SHALL be null in state

### Requirement: Workflow YAML semantic equality (REQ-014)

The `workflow_configuration_yaml` attribute SHALL use a normalized YAML type that compares values semantically (by parsing YAML to canonical JSON with sorted keys) rather than by string equality, so that formatting and key-order differences do not produce spurious plan diffs.

#### Scenario: Equivalent YAML with different formatting

- GIVEN the stored and new workflow YAML are semantically equivalent but differ in formatting or key order
- WHEN Terraform plans
- THEN no diff SHALL be produced for `workflow_configuration_yaml`

### Requirement: Version compatibility — tool export (REQ-015)

The data source SHALL verify the Kibana server version is at least 9.3.0 before performing any read operation. If the server version is below 9.3.0, the data source SHALL fail with an "Unsupported server version" error.

#### Scenario: Kibana below 9.3.0

- GIVEN the Kibana server is below version 9.3.0
- WHEN the data source reads
- THEN the provider SHALL return an "Unsupported server version" error diagnostic

### Requirement: Version compatibility — workflow export (REQ-016)

When `include_workflow` is true, the data source SHALL verify the Kibana server version is at least 9.4.0 before fetching the workflow. If the server version is below 9.4.0, the data source SHALL fail with an "Unsupported server version" error.

#### Scenario: Kibana below 9.4.0 with include_workflow

- GIVEN the Kibana server is below version 9.4.0 and `include_workflow = true`
- WHEN the data source reads
- THEN the provider SHALL return an "Unsupported server version" error diagnostic before attempting the workflow fetch
