# `elasticstack_kibana_agentbuilder_skill` (data source) — Schema and Functional Requirements

Data source implementation: `internal/kibana/agentbuilderskill`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_agentbuilder_skill` data source, which exports an existing Kibana Agent Builder skill by ID, optionally scoped to a Kibana space or accessed via an entity-local connection. Includes composite ID parsing, space resolution, version gating, and stable mapping from the API response into Terraform state.

## Schema

```hcl
data "elasticstack_kibana_agentbuilder_skill" "example" {
  skill_id = <required, string>           # bare skill id or composite "<space_id>/<skill_id>"
  space_id = <optional, computed, string> # overrides composite id space; defaults to "default"

  # computed outputs
  id          = <computed, string>        # "<space_id>/<skill_id>"
  name        = <computed, string>
  description = <computed, string>
  content     = <computed, string>
  tool_ids    = <computed, set(string)>
  referenced_content = <computed, list(object)>
    name          = <computed, string>
    relative_path = <computed, string>
    content       = <computed, string>

  kibana_connection = <optional, block>   # entity-local Kibana connection override
}
```

Notes:

- `skill_id` accepts a bare id (`my-skill`) or a composite `<space_id>/<skill_id>` string. When a composite id is supplied and `space_id` is also set explicitly, `space_id` takes precedence.
- All other attributes are read-only computed outputs populated from the API response.
- The data source enforces a minimum Elastic Stack version of `9.4.0-SNAPSHOT`.

## Requirements

### Requirement: Kibana Agent Builder Skills read API (REQ-001)

The data source SHALL read a skill using `GET /api/agent_builder/skills/{skillId}` and SHALL populate all computed attributes from the response.

#### Scenario: Successful skill read

- GIVEN a valid `skill_id` and an accessible Kibana instance
- WHEN the data source reads
- THEN all computed attributes SHALL be populated from the API response

### Requirement: API and client error surfacing (REQ-002)

When the provider cannot obtain the Kibana OpenAPI client, the read SHALL fail with an error diagnostic. Transport errors and unexpected HTTP statuses from the Skills API SHALL be surfaced as error diagnostics.

#### Scenario: Missing Kibana client

- GIVEN the data source cannot obtain a Kibana OpenAPI client
- WHEN read runs
- THEN the operation SHALL fail with an error diagnostic

### Requirement: Stack version gate (REQ-003)

Before reading, the data source SHALL verify that the target Elastic Stack version is at least `9.4.0-SNAPSHOT` via `GetVersionRequirements`. If the stack version is lower, the read SHALL fail with an `Unsupported server version` diagnostic.

#### Scenario: Stack below minimum version

- GIVEN a target Elastic Stack version below `9.4.0-SNAPSHOT`
- WHEN the data source read runs
- THEN the operation SHALL fail with an unsupported-version diagnostic

### Requirement: Provider-level Kibana client by default with scoped override (REQ-004)

The data source SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all Skills API calls.

#### Scenario: Scoped Kibana connection

- WHEN `kibana_connection` is configured on the data source
- THEN all Skills API calls SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Space and skill ID resolution (REQ-005)

The data source SHALL resolve the effective `space_id` and `skill_id` as follows: if `skill_id` parses as a composite `<space_id>/<skill_id>` string, the embedded space is used as the default space; if `space_id` is also explicitly provided, `space_id` SHALL override the composite-embedded space. If neither provides a space, the data source SHALL default to `default`. The data source SHALL normalize `skill_id` in state to the bare resource id regardless of whether a composite was supplied.

#### Scenario: Bare skill id defaults to default space

- GIVEN `skill_id` is a bare id and `space_id` is not set
- WHEN read resolves identifiers
- THEN the effective space SHALL be `default`

#### Scenario: Composite skill id extracts space

- GIVEN `skill_id` is `observability/my-skill` and `space_id` is not set
- WHEN read resolves identifiers
- THEN the effective space SHALL be `observability` and the effective skill id SHALL be `my-skill`

#### Scenario: Explicit space_id overrides composite

- GIVEN `skill_id` is `observability/my-skill` and `space_id` is `platform`
- WHEN read resolves identifiers
- THEN the effective space SHALL be `platform`

#### Scenario: skill_id normalized to bare id in state

- GIVEN `skill_id` is supplied as `default/my-skill`
- WHEN read completes
- THEN `skill_id` in state SHALL be `my-skill`

### Requirement: Not-found behavior (REQ-006)

If the skill does not exist in Kibana, the data source SHALL fail with a `Skill not found` error diagnostic rather than silently returning empty state.

#### Scenario: Skill does not exist

- GIVEN a `skill_id` that does not correspond to any skill in Kibana
- WHEN read runs
- THEN the data source SHALL fail with a `Skill not found` error diagnostic

### Requirement: State mapping from skill response (REQ-007)

When the data source maps a skill response into Terraform state, it SHALL populate `id` (`<space_id>/<skill_id>`), `skill_id`, `space_id`, `name`, `description`, `content`, `tool_ids`, and `referenced_content` from the API response. When `tool_ids` is empty or absent, it SHALL be stored as null. When `referenced_content` is empty or absent, it SHALL be stored as null. The `relative_path` field in each `referenced_content` entry SHALL be stored exactly as returned by the API.

#### Scenario: Empty tool_ids becomes null

- GIVEN a skill response that omits `tool_ids` or returns an empty array
- WHEN the provider writes state
- THEN `tool_ids` SHALL be null

#### Scenario: Canonical composite id

- GIVEN a skill with id `my-skill` read from space `observability`
- WHEN the provider writes state
- THEN `id` SHALL be `observability/my-skill`

## Traceability

| Area | Primary files |
|------|----------------|
| Schema and read callback | `internal/kibana/agentbuilderskill/data_source.go` |
| Model mapping | `internal/kibana/agentbuilderskill/models.go` |
| API client wrapper | `internal/clients/kibanaoapi/skills.go` |
| Domain models | `internal/models/agent_builder.go` |
| Composite id parsing | `internal/clients/api_client.go` |
| Space-aware routing | `internal/clients/kibanautil/spaces.go` |
| Data source envelope | `internal/entitycore/kibana_datasource.go` |
