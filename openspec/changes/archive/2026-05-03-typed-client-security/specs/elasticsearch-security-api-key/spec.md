## ADDED Requirements

### Requirement: Typed client implementation for security API key
The `elasticstack_elasticsearch_security_api_key` resource SHALL create, read, update, and invalidate API keys using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Security.CreateApiKey`, `Security.GetApiKey`, `Security.UpdateApiKey`, `Security.InvalidateApiKey`) instead of the raw `esapi` client. The typed API responses SHALL be used directly without manual JSON decoding into intermediate `models.APIKey*` types.

#### Scenario: Typed API success for API key resource create
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource creates a regular API key
- **THEN** the provider SHALL call `Security.CreateApiKey` on the typed client
- **AND** the response SHALL be returned as `*createapikey.Response`

#### Scenario: Typed API success for cross-cluster API key create
- **GIVEN** a valid Elasticsearch connection and `type = "cross_cluster"`
- **WHEN** the resource creates a cross-cluster API key
- **THEN** the provider SHALL call `Security.CreateCrossClusterApiKey` on the typed client
- **AND** the response SHALL be returned as `*createcrossclusterapikey.Response`

#### Scenario: Typed API success for API key read
- **GIVEN** a valid Elasticsearch connection
- **WHEN** the resource refreshes state
- **THEN** the provider SHALL call `Security.GetApiKey` on the typed client
- **AND** the response SHALL be used as `*types.ApiKey`
