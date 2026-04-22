## 1. Update Elasticsearch connection requirements

- [x] 1.1 Sync the `provider-elasticsearch-connection` canonical spec from the approved delta so it requires non-deprecated entity `elasticsearch_connection` schemas for SDK and Plugin Framework coverage.
- [x] 1.2 Confirm the updated requirement language still preserves helper-based source-of-truth coverage for covered Elasticsearch entities.

## 2. Remove entity deprecation metadata

- [x] 2.1 Update `internal/schema/connection.go` so entity-facing `GetEsConnectionSchema("elasticsearch_connection", false)` no longer marks the schema as deprecated.
- [x] 2.2 Update `internal/schema/connection.go` so entity-facing `GetEsFWConnectionBlock(false)` no longer exposes a deprecation message.

## 3. Extend automated verification

- [x] 3.1 Update SDK connection coverage tests to assert both helper equality and the absence of a deprecation warning for each covered entity.
- [x] 3.2 Update Plugin Framework connection coverage tests to assert both helper equality and the absence of a deprecation message for each covered entity.
- [x] 3.3 Run the relevant OpenSpec validation and targeted test commands to verify the updated requirements and helper behavior.
