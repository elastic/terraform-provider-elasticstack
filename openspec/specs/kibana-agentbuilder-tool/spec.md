# `elasticstack_kibana_agentbuilder_tool` — Schema and Functional Requirements

Resource implementation: `internal/kibana/agentbuildertool`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_tool` resource, which manages Agent Builder tools in Kibana. Tools can be of type `esql`, `index_search`, `workflow`, or `mcp`, and are scoped to a Kibana space. The resource supports create, read, update, delete, and import operations.

## Schema

```hcl
resource "elasticstack_kibana_agentbuilder_tool" "example" {
  tool_id       = <required, string>           # immutable; forces replacement on change
  type          = <required, string>           # immutable; one of: esql, index_search, workflow, mcp; forces replacement on change
  configuration = <required, string>           # JSON-encoded tool configuration; use jsonencode()

  description = <optional, string>
  tags        = <optional, list(string)>
  space_id    = <optional, computed, string>   # immutable; defaults to "default"; forces replacement on change

  # Computed
  id = <computed, string>                      # composite: "<space_id>/<tool_id>"
}
```

## Requirements

### Requirement: CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Kibana Agent Builder Tools API to manage tools. On create, the resource SHALL call `POST /api/agentbuilder/tools`. On read, the resource SHALL call `GET /api/agentbuilder/tools/{toolId}`. On update, the resource SHALL call `PUT /api/agentbuilder/tools/{toolId}`. On delete, the resource SHALL call `DELETE /api/agentbuilder/tools/{toolId}`. All API calls SHALL be scoped to the configured Kibana space via the space-aware path request editor.

#### Scenario: Lifecycle uses documented APIs

- GIVEN a tool managed by this resource
- WHEN create, read, update, or delete runs
- THEN the provider SHALL call the appropriate Agent Builder Tools API endpoint scoped to the configured space

### Requirement: Post-write read (REQ-005)

After a successful create or update, the resource SHALL immediately read the tool back from the API and use the response to populate state.

#### Scenario: State after create

- GIVEN a successful POST response
- WHEN the provider refreshes state after create
- THEN it SHALL call GET for the created tool and populate state from the response

### Requirement: Not found on read (REQ-006)

When the API returns HTTP 404 during a read (refresh), the resource SHALL remove itself from Terraform state rather than returning an error.

#### Scenario: Tool deleted outside Terraform

- GIVEN a tool no longer exists in Kibana
- WHEN refresh runs
- THEN the resource SHALL be removed from state with no error diagnostic

### Requirement: Idempotent delete (REQ-007)

When the API returns HTTP 404 during delete, the resource SHALL treat it as a successful deletion and SHALL NOT return an error.

#### Scenario: Delete already-deleted tool

- GIVEN a tool has already been deleted outside Terraform
- WHEN destroy runs
- THEN the provider SHALL return no error

### Requirement: API error surfacing (REQ-008)

When the API returns a non-success status (other than 404 on read or delete), the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Kibana API returns an error on create, update, read, or delete
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-009)

The resource SHALL expose a computed `id` in the format `<space_id>/<tool_id>`. The `id` SHALL be set after create and preserved across updates.

#### Scenario: Computed id after apply

- GIVEN a successful create
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<tool_id>` for the configured space and tool

### Requirement: Import (REQ-010)

The resource SHALL support import using the composite `id` in the format `<space_id>/<tool_id>`. On import, the provider SHALL parse the id to determine the space and tool, then read the tool from the API to populate state.

#### Scenario: Import by composite id

- GIVEN an `id` of the form `<space_id>/<tool_id>`
- WHEN import runs
- THEN the provider SHALL read the tool and populate all attributes in state

### Requirement: Immutable fields (REQ-011)

When `tool_id`, `type`, or `space_id` changes, the resource SHALL require replacement (destroy and recreate) rather than performing an in-place update.

#### Scenario: Changing tool_id

- GIVEN a configuration change to `tool_id`, `type`, or `space_id`
- WHEN Terraform plans the change
- THEN the resource SHALL be marked for replacement

### Requirement: space_id default (REQ-012)

When `space_id` is not configured, the resource SHALL default to `"default"` and SHALL preserve that value in state across plan and apply cycles.

#### Scenario: Omitted space_id

- GIVEN `space_id` is not set in configuration
- WHEN the resource is created
- THEN `space_id` in state SHALL be `"default"`

### Requirement: Configuration JSON mapping (REQ-013)

The resource SHALL accept `configuration` as a JSON-encoded string and SHALL unmarshal it into a `map[string]any` before sending to the API. When reading from the API, the resource SHALL marshal the configuration map back to a JSON string for state.

#### Scenario: Invalid configuration JSON

- GIVEN `configuration` contains invalid JSON
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL NOT call the API

### Requirement: Optional fields mapping (REQ-014)

When `description` is null or not configured, the resource SHALL omit it from API requests. When `tags` is null or empty, the resource SHALL omit it from API requests. When the API returns a null `description` or empty `tags`, the resource SHALL store null for those attributes in state.

#### Scenario: Absent optional fields

- GIVEN `description` and `tags` are not set
- WHEN the resource is created
- THEN those fields SHALL be omitted from the API request and stored as null in state

### Requirement: Version compatibility (REQ-015)

The resource SHALL verify the Kibana server version is at least 9.3.0 before performing any API operation. If the server version is below 9.3.0, the resource SHALL fail with an "Unsupported server version" error.

#### Scenario: Older Kibana version

- GIVEN the Kibana server is below version 9.3.0
- WHEN any CRUD operation runs
- THEN the provider SHALL return an "Unsupported server version" error diagnostic
