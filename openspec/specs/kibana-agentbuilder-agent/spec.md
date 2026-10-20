# `elasticstack_kibana_agentbuilder_agent` — Schema and Functional Requirements

Resource implementation: `internal/kibana/agentbuilderagent`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_agent` resource, which manages Agent Builder agents in Kibana. An agent defines a conversational AI assistant with a name, optional description, avatar, labels, tools, and system instructions. The resource supports create, read, update, delete, and import operations.

## Schema

```hcl
resource "elasticstack_kibana_agentbuilder_agent" "example" {
  agent_id = <required, string>              # immutable; forces replacement on change
  name     = <required, string>

  description   = <optional, string>
  avatar_color  = <optional, computed, string>  # hex color code (e.g. "#BFDBFF"); set "" to clear
  avatar_symbol = <optional, computed, string>  # symbol/initials (e.g. "SA"); set "" to clear
  labels        = <optional, set(string)>
  tools         = <optional, set(string)>       # tool IDs the agent can use
  skill_ids     = <optional, set(string)>       # skill IDs assigned to the agent; requires Elastic Stack 9.4.0+
  instructions  = <optional, string>            # system instructions
  access_control = {                            # optional, computed, single nested object; requires Elastic Stack 9.5.0+
    access_mode = <required, string>            # one of "public", "shared", "private"
  }

  space_id = <optional, computed, string>       # immutable; defaults to "default"; forces replacement on change

  # Computed
  id = <computed, string>                       # composite: "<space_id>/<agent_id>"
}
```

## Requirements

### Requirement: CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Kibana Agent Builder Agents API to manage agents. On create, the resource SHALL call `POST /api/agentbuilder/agents`. On read, the resource SHALL call `GET /api/agentbuilder/agents/{agentId}`. On update, the resource SHALL call `PUT /api/agentbuilder/agents/{agentId}`. On delete, the resource SHALL call `DELETE /api/agentbuilder/agents/{agentId}`. All API calls SHALL be scoped to the configured Kibana space via the space-aware path request editor.

#### Scenario: Lifecycle uses documented APIs

- GIVEN an agent managed by this resource
- WHEN create, read, update, or delete runs
- THEN the provider SHALL call the appropriate Agent Builder Agents API endpoint scoped to the configured space

### Requirement: Post-write read (REQ-005)

After a successful create or update, the resource SHALL immediately read the agent back from the API and use the response to populate state.

#### Scenario: State after create

- GIVEN a successful POST response
- WHEN the provider refreshes state after create
- THEN it SHALL call GET for the created agent and populate state from the response

### Requirement: Not found on read (REQ-006)

When the API returns HTTP 404 during a read (refresh), the resource SHALL remove itself from Terraform state rather than returning an error.

#### Scenario: Agent deleted outside Terraform

- GIVEN an agent no longer exists in Kibana
- WHEN refresh runs
- THEN the resource SHALL be removed from state with no error diagnostic

### Requirement: Idempotent delete (REQ-007)

When the API returns HTTP 404 during delete, the resource SHALL treat it as a successful deletion and SHALL NOT return an error.

#### Scenario: Delete already-deleted agent

- GIVEN an agent has already been deleted outside Terraform
- WHEN destroy runs
- THEN the provider SHALL return no error

### Requirement: API error surfacing (REQ-008)

When the API returns a non-success status (other than 404 on read or delete), the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Kibana API returns an error on create, update, read, or delete
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-009)

The resource SHALL expose a computed `id` in the format `<space_id>/<agent_id>`. The `id` SHALL be set after create and preserved across updates.

#### Scenario: Computed id after apply

- GIVEN a successful create
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<agent_id>` for the configured space and agent

### Requirement: Import (REQ-010)

The resource SHALL support import using the composite `id` in the format `<space_id>/<agent_id>`. On import, the provider SHALL parse the id to determine the space and agent, then read the agent from the API to populate state.

#### Scenario: Import by composite id

- GIVEN an `id` of the form `<space_id>/<agent_id>`
- WHEN import runs
- THEN the provider SHALL read the agent and populate all attributes in state

### Requirement: Immutable fields (REQ-011)

When `agent_id` or `space_id` changes, the resource SHALL require replacement (destroy and recreate) rather than performing an in-place update.

#### Scenario: Changing agent_id

- GIVEN a configuration change to `agent_id` or `space_id`
- WHEN Terraform plans the change
- THEN the resource SHALL be marked for replacement

### Requirement: space_id default (REQ-012)

When `space_id` is not configured, the resource SHALL default to `"default"` and SHALL preserve that value in state across plan and apply cycles.

#### Scenario: Omitted space_id

- GIVEN `space_id` is not set in configuration
- WHEN the resource is created
- THEN `space_id` in state SHALL be `"default"`

### Requirement: Avatar fields are optional and computed (REQ-013)

The resource SHALL treat `avatar_color` and `avatar_symbol` as optional fields that are also computed. When omitted from configuration, the resource SHALL leave the server value unchanged on update. When explicitly set to an empty string (`""`), the resource SHALL send an empty string to the API to clear the avatar value. After reading from the API, the resource SHALL preserve empty strings in state (not convert them to null).

#### Scenario: Setting avatar values

- GIVEN `avatar_color = "#BFDBFF"` and `avatar_symbol = "TA"` are configured
- WHEN the resource is created or updated
- THEN the API SHALL receive those values and they SHALL be stored in state

#### Scenario: Omitting avatar values leaves server values intact

- GIVEN an agent exists with `avatar_color = "#BFDBFF"` and `avatar_symbol = "TA"`
- WHEN the resource is updated without `avatar_color` or `avatar_symbol` in configuration
- THEN the API SHALL NOT receive avatar fields (preserving server values) and state SHALL match the server response

#### Scenario: Clearing avatar values explicitly

- GIVEN an agent exists with avatar values set
- WHEN the resource is updated with `avatar_color = ""` and `avatar_symbol = ""`
- THEN the API SHALL receive empty strings, clearing the avatars, and state SHALL reflect empty strings

### Requirement: Optional fields mapping (REQ-014)

When `description`, `instructions`, `labels`, or `tools` are null or not configured, the resource SHALL omit them from create API requests. On update, when these fields are null or not configured, the resource SHALL omit them from the update request body, leaving server values unchanged. When the API returns null or empty values for these fields, the resource SHALL store null for string attributes and a null set for set attributes in state.

#### Scenario: Absent optional fields on create

- GIVEN `description`, `instructions`, `labels`, and `tools` are not set
- WHEN the resource is created
- THEN those fields SHALL be omitted from the API request and stored as null in state

#### Scenario: Omitting optional fields on update leaves server values intact

- GIVEN an agent exists with `description` and `labels` set
- WHEN the resource is updated without `description` or `labels` in configuration
- THEN the API SHALL NOT receive those fields and server values SHALL be preserved

### Requirement: Tool IDs mapping (REQ-015)

The resource SHALL accept `tools` as a set of tool ID strings. On create and update, the resource SHALL wrap the tool IDs in the API's `tools` configuration structure. On read, the resource SHALL extract tool IDs from the agent's configuration and store them as a set in state.

#### Scenario: Tools configured

- GIVEN `tools = ["tool-a", "tool-b"]`
- WHEN the resource is created or updated
- THEN the API request SHALL contain those tool IDs in the agent configuration's tools array

### Requirement: Skill IDs mapping (REQ-016)

The resource SHALL accept `skill_ids` as an optional set of skill ID strings. On create, when `skill_ids` is non-empty, the resource SHALL include it in the `configuration.skill_ids` field of the POST request body; when empty or null it SHALL be omitted. On update, when `skill_ids` is non-empty the resource SHALL include it in `configuration.skill_ids`; when empty or null, the resource SHALL omit it from the update body, leaving the server value unchanged. On read, the resource SHALL populate `skill_ids` from `configuration.skill_ids` in the API response; when absent or empty, `skill_ids` SHALL be stored as null in state. `skill_ids` requires Elastic Stack 9.4.0 or later; when the stack is below that version, the field should be left unset.

#### Scenario: skill_ids on create

- GIVEN `skill_ids = ["my-skill"]` is configured
- WHEN the resource is created
- THEN the POST request SHALL include `configuration.skill_ids: ["my-skill"]`

#### Scenario: skill_ids cleared in state when absent from API response

- GIVEN the API response omits `configuration.skill_ids` or returns an empty array
- WHEN the provider writes Terraform state
- THEN `skill_ids` SHALL be null

### Requirement: Version compatibility (REQ-017)

The resource SHALL verify the Kibana server version is at least 9.3.0 before performing any API operation. If the server version is below 9.3.0, the resource SHALL fail with an "Unsupported server version" error.

#### Scenario: Older Kibana version

- GIVEN the Kibana server is below version 9.3.0
- WHEN any CRUD operation runs
- THEN the provider SHALL return an "Unsupported server version" error diagnostic

### Requirement: Resource-level Kibana connection override (REQ-018)

The resource SHALL support a `kibana_connection` block that overrides the provider's default Kibana client for that resource instance. When the block is present, the resource SHALL use the configured connection for all API calls.

#### Scenario: Resource-level connection

- GIVEN a `kibana_connection` block is configured on the resource
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the connection defined in the block instead of the provider default

### Requirement: access_control schema

The resource SHALL expose an optional, computed `access_control` single-nested attribute. When present, its nested `access_mode` attribute SHALL be required and SHALL accept only `"public"`, `"shared"`, or `"private"`.

#### Scenario: access_control attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `access_control` attribute exposing a required `access_mode` string

#### Scenario: Invalid access_mode rejected

- **GIVEN** `access_control = { access_mode = "invalid" }` is configured
- **WHEN** Terraform validates the configuration
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: access_control with no access_mode rejected

- **GIVEN** `access_control = {}` is configured
- **WHEN** Terraform validates the configuration
- **THEN** Terraform SHALL report a validation error before any API call, since `access_mode` is required within the object

### Requirement: access_control version gate

An agent create or update that sends `access_control` SHALL require Elastic Stack 9.5.0 or later. If the server is below 9.5.0, the provider SHALL return a version-requirement error diagnostic before the write API call. This attribute-specific requirement SHALL NOT block refresh or deletion, even if state carries a prior `access_control` value; the existing 9.3.0 lifecycle baseline continues to apply. Omitting `access_control` from a write SHALL be unaffected by this gate on any Elastic Stack version supported by this resource (9.3.0+).

#### Scenario: access_control sent to an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** create or update would send `access_control`
- **THEN** the provider SHALL return a version-requirement error diagnostic naming the 9.5.0 minimum
- **AND** the provider SHALL NOT perform the write API call

#### Scenario: Refresh and deletion with carried-forward mode against an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **AND** state contains an `access_control` value from a prior operation against a supported server
- **WHEN** refresh and deletion run
- **THEN** refresh SHALL proceed without a version-requirement error and leave `access_control` null in state
- **AND** deletion SHALL proceed subject only to the existing lifecycle baseline

#### Scenario: access_control omitted from a write against an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** a create or update does not send `access_control`
- **THEN** the provider SHALL proceed without an attribute-specific version-requirement error

### Requirement: access_control create mapping

On create, when `access_control` is known and non-null on the plan, the resource SHALL include `access_control.access_mode` in the `POST` request body; when null or unknown, it SHALL be omitted, leaving the server to apply its own default.

#### Scenario: access_control set on create

- **GIVEN** `access_control = { access_mode = "private" }` is configured
- **WHEN** the resource is created
- **THEN** the `POST` request body SHALL include `access_control.access_mode = "private"`

#### Scenario: access_control omitted on create

- **GIVEN** `access_control` is not configured
- **WHEN** the resource is created
- **THEN** the `POST` request body SHALL NOT include `access_control`

### Requirement: access_control update mapping

On update, the resource SHALL include `access_control.access_mode` in the `PUT` request body only when `access_control` is non-null in the **Terraform configuration** (not merely on the plan). When `access_control` is absent from configuration, the resource SHALL omit `access_control` from the update body entirely, regardless of any computed value carried forward in state or plan.

#### Scenario: access_control changed on update

- **GIVEN** an agent exists with `access_control.access_mode = "private"`
- **WHEN** the resource is updated with `access_control = { access_mode = "shared" }`
- **THEN** the `PUT` request body SHALL include `access_control.access_mode = "shared"`

#### Scenario: access_control removed from config on update relinquishes management

- **GIVEN** an agent exists with `access_control.access_mode = "shared"` previously configured and present in state
- **AND** another mutable attribute is changed so that an update is required
- **WHEN** the resource is updated with `access_control` removed from configuration
- **THEN** the `PUT` request body SHALL NOT include `access_control`
- **AND** the plan and prior state SHALL still contain the carried-forward mode before the update
- **AND** a subsequent read SHALL report whatever `access_mode` the server currently has, without the provider asserting a client-side reset

### Requirement: access_control read-back via dedicated endpoint

After a successful create, update, or refresh (read) of an agent — and only when the configured Kibana server supports the Agent Builder access-control sub-resource (9.5.0+) — the resource SHALL call the dedicated `GET /api/agent_builder/agents/{id}/access_control` endpoint and populate `access_control.access_mode` in state from its response. When the server is below that version, `access_control` SHALL be left null in state and this additional call SHALL NOT be attempted.

#### Scenario: State populated from dedicated endpoint on supported stack

- **GIVEN** the Kibana server is 9.5.0 or later and an agent exists with an assigned `access_mode`
- **WHEN** the resource is created, updated, or refreshed
- **THEN** the provider SHALL call the dedicated access-control `GET` endpoint and `access_control.access_mode` in state SHALL match its response

#### Scenario: access_control left null on unsupported stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** the resource is created, updated, or refreshed
- **THEN** `access_control` SHALL be null in state and the provider SHALL NOT call the dedicated access-control endpoint

#### Scenario: Configured mode drift detected when changed out-of-band

- **GIVEN** the Kibana server is 9.5.0 or later and an agent has an explicitly configured `access_mode`
- **AND** its server mode was changed outside Terraform to a different value since the last apply
- **WHEN** `terraform plan` runs
- **THEN** the plan SHALL show a diff on `access_control.access_mode` to restore the explicitly configured value

#### Scenario: Unmanaged mode adopted on refresh

- **GIVEN** the Kibana server is 9.5.0 or later and `access_control` is omitted from configuration or management has been relinquished
- **AND** the server has an `access_mode` value
- **WHEN** refresh runs
- **THEN** state SHALL adopt the server's current `access_mode`
- **AND** a subsequent plan SHALL NOT propose a corrective update for that mode

### Requirement: No read-back for a missing agent

The dedicated access-control read-back SHALL NOT be attempted when the main agent `GET` for the same refresh returned HTTP 404 (agent not found); the existing not-found-removes-from-state handling SHALL take precedence.

#### Scenario: No read-back attempted for a deleted agent

- **GIVEN** an agent no longer exists in Kibana
- **WHEN** refresh runs
- **THEN** the provider SHALL remove the resource from state per the existing not-found handling and SHALL NOT call the dedicated access-control endpoint
