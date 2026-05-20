# `elasticstack_kibana_agentbuilder_skill` — Schema and Functional Requirements

Resource implementation: `internal/kibana/agentbuilderskill`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_skill` resource, including Kibana Agent Builder Skills API usage, composite identity and import behavior, stack-version gating, replacement behavior for immutable identifiers, `relative_path` validation, and stable mapping from API skill responses into Terraform state.

## Schema

```hcl
resource "elasticstack_kibana_agentbuilder_skill" "example" {
  id       = <computed, string>          # canonical state id: "<space_id>/<skill_id>"; UseStateForUnknown
  skill_id = <required, string>          # RequiresReplace; minimum length 1
  space_id = <optional, computed, string> # default "default"; RequiresReplace; UseStateForUnknown

  name        = <required, string>       # minimum length 1
  description = <required, string>
  content     = <required, string>       # markdown instructions

  tool_ids = <optional, set(string)>     # tool registry IDs this skill references

  referenced_content = <optional, list(object)> # ordered; up to 100 entries
    name          = <required, string>
    relative_path = <required, string>   # must match ^\./ (e.g. "./runbooks/guide.md")
    content       = <required, string>

  kibana_connection = <optional, block>  # entity-local Kibana connection override
}
```

Notes:

- `skill_id` and `space_id` are immutable identifiers; changes require resource replacement.
- `relative_path` values in `referenced_content` must start with `./`; the resource enforces this with a schema validator so the constraint is caught at plan time.
- The `referenced_content` list preserves order as provided; items are sent to and received from the API in that order.
- `tool_ids` is stored as a set (order-independent).
- CRUD operations enforce a minimum Elastic Stack version of `9.4.0-SNAPSHOT`.

## Requirements

### Requirement: Kibana Agent Builder Skills APIs (REQ-001)

The resource SHALL manage skills through the Kibana Agent Builder Skills API: `POST /api/agent_builder/skills` for create, `GET /api/agent_builder/skills/{skillId}` for read, `PUT /api/agent_builder/skills/{skillId}` for update, and `DELETE /api/agent_builder/skills/{skillId}` for delete. After a successful create or update call, the resource SHALL perform a follow-up GET request and SHALL use that GET response as the authoritative source for Terraform state.

#### Scenario: Create then authoritative read

- GIVEN a successful skill create request
- WHEN the resource completes create
- THEN it SHALL read the created skill from Kibana and SHALL write state from that read response

#### Scenario: Update then authoritative read

- GIVEN a successful skill update request
- WHEN the resource completes update
- THEN it SHALL read the skill from Kibana again and SHALL write state from that read response

### Requirement: API and client error surfacing (REQ-002)

For create, read, update, and delete, when the provider cannot obtain the Kibana OpenAPI client, the operation SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses from the Skills API SHALL be surfaced as error diagnostics, except that delete not-found (HTTP 404) SHALL be treated as success and read not-found SHALL remove the resource from state (see REQ-010).

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Delete of an already absent skill

- GIVEN a delete request for a skill that no longer exists in Kibana
- WHEN Kibana returns HTTP 404 to the delete operation
- THEN the resource SHALL treat delete as successful

#### Scenario: Delete blocked by referencing agents

- GIVEN a delete request for a skill that is still referenced by one or more agents
- WHEN Kibana returns HTTP 409 to the delete operation
- THEN the resource SHALL surface an error diagnostic describing the conflict

### Requirement: Stack version gate (REQ-003)

Before create, read, update, and delete, the resource SHALL verify that the target Elastic Stack version is at least `9.4.0-SNAPSHOT`. If the stack version is lower, the resource SHALL fail the operation with an `Unsupported server version` diagnostic stating that Agent Builder skills require that minimum version or later.

#### Scenario: Stack below minimum version

- GIVEN a target Elastic Stack version below `9.4.0-SNAPSHOT`
- WHEN any skill resource operation runs
- THEN the operation SHALL fail with an unsupported-version diagnostic before calling the Skills API

### Requirement: Provider-level Kibana client by default with scoped override (REQ-004)

The resource SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all skill operations.

#### Scenario: Standard provider configuration

- WHEN `kibana_connection` is not configured on the resource
- THEN all Skills API calls SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- WHEN `kibana_connection` is configured on the resource
- THEN all Skills API calls SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Space-aware API routing (REQ-005)

All Skills API calls SHALL be routed through the Kibana space identified by `space_id` using a space-aware path request editor. This transparently prefixes the API path with `/s/{space_id}/` for non-default spaces and uses the root path for `default`.

#### Scenario: Non-default space routing

- GIVEN `space_id` is set to `observability`
- WHEN any Skills API call is made
- THEN the request URL SHALL include `/s/observability/` in the path

### Requirement: Identity and canonical `id` (REQ-006)

The resource SHALL expose a computed canonical `id` in the format `<space_id>/<skill_id>`. After create, the resource SHALL derive `id` from `space_id` and the skill id returned by the API. After read or update, the resource SHALL derive `id` and `space_id` from the composite id already in state.

#### Scenario: Canonical composite id in state

- GIVEN a skill created in space `default` with skill id `my-skill`
- WHEN Terraform state is written
- THEN `id` SHALL be `default/my-skill` and `skill_id` SHALL be `my-skill`

### Requirement: Import passthrough and composite id expectations (REQ-007)

The resource SHALL support Terraform import by passing the supplied import identifier directly into the `id` attribute. A functional imported id MUST be in the format `<space_id>/<skill_id>` with exactly one `/`; if later operations parse a different format, they SHALL return the composite-id diagnostic produced by the shared composite-id parser.

#### Scenario: Composite import id

- GIVEN an import id `default/my-skill`
- WHEN import runs
- THEN the resource SHALL persist that string to `id` for subsequent refresh

#### Scenario: Non-composite imported id

- GIVEN an imported `id` that does not contain exactly one `/`
- WHEN a later read, update, or delete parses the id
- THEN the resource SHALL return an error diagnostic describing the required composite id format

### Requirement: Lifecycle, defaults, and unknown preservation (REQ-008)

If `space_id` is omitted, the resource SHALL default it to `default`. Changes to `skill_id` or `space_id` SHALL require resource replacement rather than in-place update. Unknown planned values for `id` and `space_id` SHALL preserve prior state through `UseStateForUnknown`.

#### Scenario: Minimal configuration defaults the space

- GIVEN configuration that omits `space_id`
- WHEN Terraform plans the resource
- THEN `space_id` SHALL default to `default`

#### Scenario: Replace on skill id change

- GIVEN an existing managed skill
- WHEN `skill_id` changes in configuration
- THEN Terraform SHALL plan resource replacement

#### Scenario: Replace on space id change

- GIVEN an existing managed skill
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan resource replacement

### Requirement: `relative_path` validator (REQ-009)

Each `relative_path` value in `referenced_content` SHALL be validated against the regular expression `^\./`. If a value does not start with `./`, Terraform validation SHALL fail with an attribute diagnostic on `relative_path` stating that it must start with `./`, and the provider SHALL not call the Skills API.

#### Scenario: Relative path without `./` prefix

- GIVEN a `referenced_content` entry whose `relative_path` is `runbooks/guide.md`
- WHEN Terraform validates the resource configuration
- THEN the resource SHALL return an attribute diagnostic on `relative_path`

#### Scenario: Valid relative path

- GIVEN a `referenced_content` entry whose `relative_path` is `./runbooks/guide.md`
- WHEN Terraform validates the resource configuration
- THEN no validation diagnostic SHALL be returned for `relative_path`

### Requirement: Create and update request mapping (REQ-010)

On create, the resource SHALL send `id`, `name`, `description`, `content`, `tool_ids` (omitted when empty), and `referenced_content` (omitted when empty) in the POST body. On update, the resource SHALL identify the target skill from the composite `id` in state and SHALL send `name`, `description`, `content`, `tool_ids` (always present; empty slice when unset), and `referenced_content` (always present; empty slice when unset) in the PUT body.

#### Scenario: Create omits empty optional fields

- GIVEN configuration that sets only `skill_id`, `name`, `description`, and `content`
- WHEN create builds the API request
- THEN the request SHALL omit `tool_ids` and `referenced_content`

#### Scenario: Update sends empty slices to clear optional fields

- GIVEN a skill whose `tool_ids` and `referenced_content` are removed from configuration
- WHEN update runs
- THEN the PUT request SHALL include `tool_ids: []` and `referenced_content: []` to clear those fields in Kibana

### Requirement: Read behavior and missing resources (REQ-011)

When refreshing state, the resource SHALL parse the composite `id` to obtain the current `space_id` and `skill_id`, SHALL restore `space_id` into the model before mapping the API response, and SHALL remove the resource from Terraform state when Kibana reports that the skill is not found.

#### Scenario: Read of deleted skill

- GIVEN a resource recorded in Terraform state
- WHEN the read operation calls Kibana and the skill is not found
- THEN the provider SHALL remove the resource from state

### Requirement: State mapping from skill responses (REQ-012)

When the resource maps a skill response into Terraform state, it SHALL populate `id`, `skill_id`, `space_id`, `name`, `description`, `content`, `tool_ids`, and `referenced_content` from the API response. When `tool_ids` is empty or absent in the response, it SHALL be stored as null. When `referenced_content` is empty or absent, it SHALL be stored as null. The `relative_path` field in each `referenced_content` entry SHALL be stored exactly as returned by the API (including the `./` prefix).

#### Scenario: Empty tool_ids becomes null

- GIVEN a skill response that omits `tool_ids` or returns an empty array
- WHEN the provider writes Terraform state
- THEN `tool_ids` SHALL be null

#### Scenario: Empty referenced_content becomes null

- GIVEN a skill response that omits `referenced_content` or returns an empty array
- WHEN the provider writes Terraform state
- THEN `referenced_content` SHALL be null

## Traceability

| Area | Primary files |
|------|----------------|
| Schema | `internal/kibana/agentbuilderskill/schema.go` |
| Metadata / Configure / Import | `internal/kibana/agentbuilderskill/resource.go` |
| CRUD orchestration | `internal/kibana/agentbuilderskill/create.go`, `internal/kibana/agentbuilderskill/read.go`, `internal/kibana/agentbuilderskill/update.go`, `internal/kibana/agentbuilderskill/delete.go` |
| Model mapping | `internal/kibana/agentbuilderskill/models.go` |
| API client wrappers | `internal/clients/kibanaoapi/skills.go` |
| Domain models | `internal/models/agent_builder.go` |
| Composite id parsing | `internal/clients/api_client.go` |
| Space-aware routing | `internal/clients/kibanautil/spaces.go` |
