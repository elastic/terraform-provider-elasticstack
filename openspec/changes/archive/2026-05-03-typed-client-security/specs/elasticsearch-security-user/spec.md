## ADDED Requirements

### Requirement: Typed client implementation for security user
The `elasticstack_elasticsearch_security_user` resource and data source SHALL retrieve and manage users using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Security.PutUser`, `Security.GetUser`, `Security.DeleteUser`) instead of the raw `esapi` client. The typed API response SHALL be used directly without manual JSON decoding into an intermediate `models.User` type.

#### Scenario: Typed API success for user resource
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource performs create, read, update, or delete
- **THEN** the provider SHALL call the typed Security user APIs
- **AND** user data SHALL be returned as `*types.User`

#### Scenario: Typed API success for user data source
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the data source reads a user
- **THEN** the provider SHALL call `Security.GetUser` on the typed client
- **AND** the response SHALL be used as `getuser.Response`
