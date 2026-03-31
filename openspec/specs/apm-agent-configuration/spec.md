# `elasticstack_apm_agent_configuration` — Schema and Functional Requirements

Resource implementation: `internal/apm/agent_configuration`

## Purpose

Define schema and behavior for the APM agent configuration resource: API usage, identity/import, connection, create/update, read, delete, and mapping between Terraform state and the Kibana APM agent configuration API.

## Schema

```hcl
resource "elasticstack_apm_agent_configuration" "example" {
  id = <computed, string> # internal identifier: <service_name> or <service_name>:<service_environment>

  service_name        = <required, string>
  service_environment = <optional, string>
  agent_name          = <optional, string>
  settings            = <required, map(string)>
}
```

## Requirements

### Requirement: APM agent configuration CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Kibana APM agent configuration API (`CreateUpdateAgentConfiguration`) to create and update agent configurations ([docs](https://www.elastic.co/docs/solutions/observability/apm/apm-agent-central-configuration)). The resource SHALL use the Kibana APM agent configuration API (`GetAgentConfigurations`) to read all agent configurations and match against state identity. The resource SHALL use the Kibana APM agent configuration API (`DeleteAgentConfiguration`) to delete agent configurations. When the Kibana API returns a non-success status for any create, update, read, or delete request, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Kibana API returns a non-success HTTP status for a create request
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error and the resource SHALL not be stored in state

#### Scenario: API failure on read

- GIVEN the Kibana API returns a non-success HTTP status for a read request
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Kibana client usage (REQ-005)

The resource SHALL obtain its Kibana client via `GetKibanaOapiClient()` from the provider-configured API client for all API operations. The resource SHALL use the Elastic API version `2023-10-31` in all API requests.

#### Scenario: Kibana client acquisition failure

- GIVEN the provider cannot provide a Kibana client
- WHEN any CRUD operation runs
- THEN Terraform diagnostics SHALL include an "Unable to get Kibana client" error

### Requirement: Identity (REQ-006–REQ-007)

The resource SHALL expose a computed `id` that is derived from the service identity. When `service_environment` is null, empty, or unknown, the `id` SHALL be set to `<service_name>`. When `service_environment` is a non-empty string, the `id` SHALL be set to `<service_name>:<service_environment>`. The `id` SHALL be preserved across plan/apply cycles using `UseStateForUnknown`.

#### Scenario: Identity without environment

- GIVEN `service_name` is "my-service" and `service_environment` is not set
- WHEN the resource is created
- THEN `id` SHALL be "my-service"

#### Scenario: Identity with environment

- GIVEN `service_name` is "my-service" and `service_environment` is "production"
- WHEN the resource is created
- THEN `id` SHALL be "my-service:production"

### Requirement: Import (REQ-008)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the imported `id` value directly to Terraform state. The accepted `id` format for import SHALL be `<service_name>` or `<service_name>:<service_environment>`.

#### Scenario: Import passthrough

- GIVEN an import with a valid `id` of `<service_name>:<service_environment>`
- WHEN import completes
- THEN the `id` SHALL be stored in state and a read SHALL be performed to populate all other attributes

### Requirement: Create (REQ-009–REQ-011)

On create, the resource SHALL read the plan and construct a `CreateUpdateAgentConfiguration` request body containing `service.name`, `service.environment`, `agent_name`, and `settings`. The resource SHALL call the create/update API without the `overwrite` flag. After a successful API response, the resource SHALL derive the `id` from the service identity and perform a read to populate the full state.

#### Scenario: Create with all fields

- GIVEN a plan with `service_name`, `service_environment`, `agent_name`, and `settings`
- WHEN create runs
- THEN the API SHALL be called with all configured fields and state SHALL be refreshed from the API after create

### Requirement: Update (REQ-012–REQ-013)

On update, the resource SHALL read the plan and construct a `CreateUpdateAgentConfiguration` request body with `service.name`, `service.environment`, `agent_name`, and `settings`. The resource SHALL call the API with `overwrite=true`. After a successful API response, the resource SHALL perform a read to refresh state.

#### Scenario: Update with overwrite

- GIVEN an existing resource with changed `settings`
- WHEN update runs
- THEN the API SHALL be called with `overwrite=true` and state SHALL be refreshed from the API after update

### Requirement: Read (REQ-014–REQ-017)

On read, the resource SHALL call the `GetAgentConfigurations` API to retrieve all configurations. The resource SHALL compute the expected `id` for each returned configuration using the same logic as `SetIDFromService` (i.e., `<service_name>` or `<service_name>:<service_environment>`) and match it against the `id` stored in state. If no configuration matches the state `id`, the resource SHALL remove itself from state. If the API returns a nil body for a 200 response, the resource SHALL return an error diagnostic.

#### Scenario: Not found removes from state

- GIVEN no APM agent configuration matches the state `id`
- WHEN read runs
- THEN the resource SHALL be removed from state without error

#### Scenario: Nil response body

- GIVEN the API returns a 200 response with a nil JSON body
- WHEN read runs
- THEN Terraform diagnostics SHALL include an error about the unexpected nil body

### Requirement: Delete (REQ-018–REQ-019)

On delete, the resource SHALL parse the state `id` by splitting on `:` to extract `service_name` (first segment) and optionally `service_environment` (second segment, if present). The resource SHALL call the `DeleteAgentConfiguration` API with a request body containing the parsed service name and, if present, the service environment.

#### Scenario: Delete without environment

- GIVEN a state `id` of "my-service" (no `:` separator)
- WHEN delete runs
- THEN the API SHALL be called with `service.name = "my-service"` and no `service.environment`

#### Scenario: Delete with environment

- GIVEN a state `id` of "my-service:production"
- WHEN delete runs
- THEN the API SHALL be called with `service.name = "my-service"` and `service.environment = "production"`

### Requirement: Settings mapping (REQ-020–REQ-021)

The `settings` attribute SHALL be a required map of string-to-string values. On create and update, the resource SHALL convert the `settings` map to `map[string]string` and pass it directly as the API `settings` field. On read, when the API returns settings, the resource SHALL convert each value to a string using `fmt.Sprintf("%v", v)` and store the resulting map in state as `map(string)`.

#### Scenario: Settings round-trip

- GIVEN `settings = { "log_level" = "debug" }` configured in plan
- WHEN create or update runs
- THEN the API SHALL receive `{ "log_level": "debug" }` and state SHALL reflect the same values after read

### Requirement: Read state mapping (REQ-022–REQ-025)

On read, after matching the configuration by `id`, the resource SHALL set `service_name` from `foundConfig.Service.Name`. The resource SHALL set `service_environment` from `foundConfig.Service.Environment` (nil pointer → null in state). The resource SHALL set `agent_name` from `foundConfig.AgentName` (nil pointer → null in state). The resource SHALL set `settings` from the API response settings map, converting all values to strings. If `foundConfig.Settings` is nil, the resource SHALL store an empty map in state for `settings`.

#### Scenario: Optional fields nil from API

- GIVEN the API returns a configuration where `agent_name` and `service_environment` are nil
- WHEN read runs
- THEN `agent_name` and `service_environment` SHALL be null in state
