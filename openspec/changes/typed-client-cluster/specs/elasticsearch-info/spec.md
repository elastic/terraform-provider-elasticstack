## ADDED Requirements

### Requirement: Typed client implementation for cluster info
The data source SHALL retrieve cluster metadata using the go-elasticsearch Typed API (`elasticsearch.TypedClient.Core.Info().Do(ctx)`) instead of the raw `esapi` client. The typed API response SHALL be returned directly without manual JSON decoding into an intermediate model type.

#### Scenario: Typed API success
- GIVEN a valid Elasticsearch connection
- WHEN the data source reads cluster info
- THEN the provider SHALL call `Core.Info().Do(ctx)` on the typed client
- AND the response SHALL be returned as `*core.InfoResponse`

#### Scenario: Build date formatting
- GIVEN a successful typed API response
- WHEN the data source maps the `version.build_date` field
- THEN the value SHALL be formatted using the typed API's `DateTime.String()` method
- AND the resulting string SHALL remain a valid timestamp representation
