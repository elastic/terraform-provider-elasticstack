## ADDED Requirements

### Requirement: Kibana OpenAPI client for Spaces (REQ-018)

The `elasticstack_kibana_space` implementation SHALL perform Create Space, Get Space, Update Space, and Delete Space HTTP calls using the generated OpenAPI Kibana client package (`generated/kbapi`) and helper functions colocated under `internal/clients/kibanaoapi` for Spaces. The implementation SHALL NOT use the legacy `go-kibana-rest` `KibanaSpaces` API methods for those operations after this migration.

#### Scenario: Mutations use kbapi transport

- **GIVEN** a create, update, read, or delete operation for the resource
- **WHEN** the provider issues the corresponding Kibana Spaces HTTP request
- **THEN** the request SHALL be executed through the kbapi client configured for the effective Kibana connection (provider default or `kibana_connection` scoped client)

### Requirement: Typed space models aligned with legacy KibanaSpace (REQ-019)

The kbapi types used to decode and encode space request and response JSON SHALL include the same logical fields as `github.com/disaster37/go-kibana-rest/v8/kbapi.KibanaSpace` for provider purposes: `id`, `name`, `description`, `disabledFeatures`, `initials`, `color`, `imageUrl`, `solution`, and `_reserved` when returned by Kibana. Schema adjustments needed to achieve this SHALL be implemented in `generated/kbapi/transform_schema.go` followed by regeneration of `generated/kbapi`.

#### Scenario: Read maps equivalent Terraform attributes

- **GIVEN** a successful Get Space response identical to one handled by the pre-migration implementation
- **WHEN** read populates Terraform state from the typed kbapi response
- **THEN** the values stored for `space_id`, `name`, `description`, `disabled_features`, `initials`, `color`, and `solution` SHALL match the values the legacy client mapping would have produced for that JSON payload

### Requirement: solution version gate uses effective connection (REQ-020)

When the `solution` argument is set to a non-empty value, the resource SHALL evaluate the minimum Kibana/Stack version using the same effective Kibana connection as the kbapi Spaces calls before performing create or update, and SHALL fail with diagnostics when the version is below 8.16.0 as specified in existing requirements.

#### Scenario: Version check precedes kbapi mutation

- **GIVEN** `solution` is set and the resolved server version is below 8.16.0
- **WHEN** create or update runs
- **THEN** the resource SHALL return an error diagnostic without issuing a successful mutating Spaces API call that would persist an unsupported `solution`
