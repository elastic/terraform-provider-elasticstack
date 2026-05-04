## ADDED Requirements

### Requirement: Typed client implementation for security role
The `elasticstack_elasticsearch_security_role` resource and data source SHALL retrieve and manage roles using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Security.PutRole`, `Security.GetRole`, `Security.DeleteRole`) instead of the raw `esapi` client. The typed API response SHALL be used directly without manual JSON decoding into an intermediate `models.Role` type.

#### Scenario: Typed API success for role resource
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource performs create, read, update, or delete
- **THEN** the provider SHALL call the typed Security role APIs
- **AND** role data SHALL be returned as `*types.Role`

#### Scenario: Typed API success for role data source
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the data source reads a role
- **THEN** the provider SHALL call `Security.GetRole` on the typed client
- **AND** the response SHALL be used as `getrole.Response`
