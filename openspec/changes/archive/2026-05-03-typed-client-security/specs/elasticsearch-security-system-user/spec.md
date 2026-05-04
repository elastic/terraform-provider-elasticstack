## ADDED Requirements

### Requirement: Typed client implementation for security system user
The `elasticstack_elasticsearch_security_system_user` resource SHALL read system users and toggle their enabled state using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Security.GetUser`, `Security.EnableUser`, `Security.DisableUser`, `Security.ChangePassword`) instead of the raw `esapi` client. The typed API response SHALL be used directly without manual JSON decoding into an intermediate `models.User` type.

#### Scenario: Typed API success for system user resource
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource performs create, read, update, or delete
- **THEN** the provider SHALL call the typed Security user APIs
- **AND** user data SHALL be returned as `*types.User`
