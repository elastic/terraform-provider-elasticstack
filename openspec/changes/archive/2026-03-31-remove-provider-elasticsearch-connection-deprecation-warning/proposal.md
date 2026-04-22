## Why

The current `provider-elasticsearch-connection` requirements enforce coverage and helper equivalence for `elasticsearch_connection`, but they do not define whether entity-level connection schemas should expose a deprecation warning. As a result, `internal/schema.GetEsConnectionSchema("elasticsearch_connection", false)` and `internal/schema.GetEsFWConnectionBlock(false)` still carry a warning that conflicts with the desired provider behavior.

## What Changes

- Update the `provider-elasticsearch-connection` capability to require that covered Elasticsearch entity schemas do not expose a deprecation warning on `elasticsearch_connection`.
- Preserve the existing helper-based source-of-truth requirements so SDK and Plugin Framework entities continue to share a single connection schema definition.
- Add acceptance criteria for SDK and Plugin Framework coverage tests that verify the absence of an entity-level deprecation warning.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `provider-elasticsearch-connection`: require covered entity `elasticsearch_connection` definitions to be non-deprecated in both SDK and Plugin Framework schemas

## Impact

- `openspec/specs/provider-elasticsearch-connection/spec.md`
- `internal/schema/connection.go`
- Provider tests that assert schema and block equivalence for `elasticsearch_connection`
