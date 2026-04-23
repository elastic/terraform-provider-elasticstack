## ADDED Requirements

### Requirement: Agent CRUD APIs (REQ-001)
The resource SHALL use the Kibana Agent Builder Agents API to manage agents. On create, the resource SHALL call the agent create operation. On read, the resource SHALL call the agent get operation. On update, the resource SHALL call the agent update operation. On delete, the resource SHALL call the agent delete operation. All API calls SHALL be scoped to the configured Kibana space. The Terraform lifecycle orchestration for these operations SHALL be performed through the shared Kibana Plugin Framework generic resource framework, while request construction, response mapping, and agent-specific transport remain resource-specific components.

#### Scenario: Lifecycle uses Agent Builder agent APIs
- **WHEN** an agent resource is created, refreshed, updated, or deleted
- **THEN** the provider SHALL call the corresponding Agent Builder Agents API in the configured Kibana space

### Requirement: Agent post-write read
After a successful create or update, the shared Kibana Plugin Framework generic resource framework SHALL immediately read the agent back from the API and SHALL use that response to populate Terraform state.

#### Scenario: State after create
- **WHEN** the provider creates an agent successfully
- **THEN** it SHALL perform an immediate follow-up read and SHALL write state from that authoritative response

### Requirement: Agent identity and import
The resource SHALL expose a computed `id` in the format `<space_id>/<agent_id>`. The resource SHALL support import using that composite identity, and shared composite-ID helpers from the Kibana Plugin Framework generic resource framework SHALL parse the stored identity for subsequent lifecycle operations.

#### Scenario: Computed id after apply
- **WHEN** a successful create populates state
- **THEN** the `id` attribute SHALL equal `<space_id>/<agent_id>`

#### Scenario: Composite import id
- **WHEN** import is invoked with `<space_id>/<agent_id>`
- **THEN** the provider SHALL store that value in `id` and SHALL use it to read the agent

### Requirement: Agent missing resource handling
When the agent read operation reports that the remote agent no longer exists, the shared Kibana Plugin Framework generic resource framework SHALL remove the resource from Terraform state rather than returning an error.

#### Scenario: Agent deleted outside Terraform
- **WHEN** refresh runs after the remote agent has been deleted out of band
- **THEN** the provider SHALL remove the resource from Terraform state

### Requirement: Agent schema defaults and replacement semantics
`space_id` SHALL default to `default` when omitted. Changes to `agent_id` or `space_id` SHALL require replacement rather than in-place update.

#### Scenario: Default space id
- **WHEN** `space_id` is omitted from configuration
- **THEN** the provider SHALL use `default` as the effective space and SHALL persist that value in state

#### Scenario: Changing immutable identity fields
- **WHEN** `agent_id` or `space_id` changes in configuration
- **THEN** Terraform SHALL plan resource replacement

### Requirement: Agent field mapping
The resource SHALL map Terraform configuration and remote API state for the following fields: `name`, `description`, `avatar_color`, `avatar_symbol`, `labels`, `tools`, and `instructions`. Null or empty optional remote values for `description`, `avatar_color`, `avatar_symbol`, and `instructions` SHALL be stored as null in Terraform state.

#### Scenario: Empty optional remote fields become null
- **WHEN** the remote agent omits or returns empty optional descriptive fields
- **THEN** the provider SHALL store those attributes as null in Terraform state

### Requirement: Agent version compatibility
The resource SHALL verify the Kibana server version is at least 9.3.0 before performing any API operation. The minimum-version requirement and user-facing unsupported-version diagnostic SHALL be supplied by the resource model and enforced by the shared Kibana Plugin Framework generic resource framework.

#### Scenario: Kibana below minimum version
- **WHEN** the Kibana server version is below 9.3.0 and an agent lifecycle operation runs
- **THEN** the provider SHALL return an `Unsupported server version` diagnostic before invoking the agent API
