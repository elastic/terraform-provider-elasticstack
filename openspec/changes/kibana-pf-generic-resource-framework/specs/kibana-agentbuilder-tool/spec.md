## MODIFIED Requirements

### Requirement: CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Kibana Agent Builder Tools API to manage tools. On create, the resource SHALL call `POST /api/agentbuilder/tools`. On read, the resource SHALL call `GET /api/agentbuilder/tools/{toolId}`. On update, the resource SHALL call `PUT /api/agentbuilder/tools/{toolId}`. On delete, the resource SHALL call `DELETE /api/agentbuilder/tools/{toolId}`. All API calls SHALL be scoped to the configured Kibana space via the space-aware path request editor. The Terraform lifecycle orchestration for these operations SHALL be performed through the shared Kibana Plugin Framework generic resource framework, while request construction, response mapping, and tool-specific transport remain resource-specific components.

#### Scenario: Lifecycle uses documented APIs

- GIVEN a tool managed by this resource
- WHEN create, read, update, or delete runs
- THEN the provider SHALL call the appropriate Agent Builder Tools API endpoint scoped to the configured space

### Requirement: Post-write read (REQ-005)

After a successful create or update, the resource SHALL immediately read the tool back from the API and use the response to populate state. The shared Kibana Plugin Framework generic resource framework SHALL perform this read-after-write orchestration.

#### Scenario: State after create

- GIVEN a successful POST response
- WHEN the provider refreshes state after create
- THEN it SHALL call GET for the created tool and populate state from the response

### Requirement: Not found on read (REQ-006)

When the API returns HTTP 404 during a read (refresh), the shared Kibana Plugin Framework generic resource framework SHALL remove the resource from Terraform state rather than returning an error.

#### Scenario: Tool deleted outside Terraform

- GIVEN a tool no longer exists in Kibana
- WHEN refresh runs
- THEN the resource SHALL be removed from state with no error diagnostic

### Requirement: Import (REQ-010)

The resource SHALL support import using the composite `id` in the format `<space_id>/<tool_id>`. Shared composite-ID helpers from the Kibana Plugin Framework generic resource framework SHALL parse the import/state identity to determine the space and tool identifiers used for subsequent reads and lifecycle operations.

#### Scenario: Import by composite id

- GIVEN an `id` of the form `<space_id>/<tool_id>`
- WHEN import runs
- THEN the provider SHALL read the tool and populate all attributes in state

### Requirement: Version compatibility (REQ-015)

The resource SHALL verify the Kibana server version is at least 9.3.0 before performing any API operation. The minimum-version requirement and user-facing unsupported-version diagnostic SHALL be supplied by the resource model and enforced by the shared Kibana Plugin Framework generic resource framework. If the server version is below 9.3.0, the resource SHALL fail with an "Unsupported server version" error.

#### Scenario: Older Kibana version

- GIVEN the Kibana server is below version 9.3.0
- WHEN any CRUD operation runs
- THEN the provider SHALL return an "Unsupported server version" error diagnostic
