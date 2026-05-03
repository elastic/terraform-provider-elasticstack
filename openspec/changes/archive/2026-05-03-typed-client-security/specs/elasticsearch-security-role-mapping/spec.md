## ADDED Requirements

### Requirement: Typed client implementation for security role mapping
The `elasticstack_elasticsearch_security_role_mapping` resource and data source SHALL retrieve and manage role mappings using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Security.PutRoleMapping`, `Security.GetRoleMapping`, `Security.DeleteRoleMapping`) instead of the raw `esapi` client. The typed API response SHALL be used directly without manual JSON decoding into an intermediate `models.RoleMapping` type.

#### Scenario: Typed API success for role mapping resource
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource performs create, read, update, or delete
- **THEN** the provider SHALL call the typed Security role mapping APIs
- **AND** role mapping data SHALL be returned as `*types.SecurityRoleMapping`

#### Scenario: Typed API success for role mapping data source
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the data source reads a role mapping
- **THEN** the provider SHALL call `Security.GetRoleMapping` on the typed client
- **AND** the response SHALL be used as `getrolemapping.Response`
