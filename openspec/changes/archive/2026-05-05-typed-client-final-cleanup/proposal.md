## Why

The typed-client migration has progressed through all prior phases and every Elasticsearch API consumer has been moved to the `go-elasticsearch` Typed API. The raw `esapi` client, bridging helpers, and hand-rolled model types are now entirely unused and represent dead code. Removing them reduces maintenance surface, eliminates confusion for future contributors, and completes the migration.

## What Changes

- **Update `ElasticsearchScopedClient` to return `*elasticsearch.TypedClient` directly**
  - Remove the raw `elasticsearch` field from `ElasticsearchScopedClient`
  - Rename `GetESTypedClient()` (added in prior phases) to `GetESClient()` and change its return type to `*elasticsearch.TypedClient`
  - Update the endpoint-validation logic to work with the typed-client accessor
  - **BREAKING** for any internal code still referencing the raw client (none expected — all consumers already migrated)

- **Delete `internal/clients/elasticsearch/helpers.go`**
  - `doFWWrite` and `doSDKWrite` are obsolete because typed API methods handle JSON marshaling and response parsing internally

- **Delete redundant model files from `internal/models/`**
  - `models.go` — remove types that have typedapi equivalents: `ClusterInfo`, `User`, `Role`, `APIKey*`, `RoleMapping`, `IndexTemplate*`, `ComponentTemplate*`, `Policy*`, `SnapshotRepository`, `SnapshotPolicy*`, `DataStream*`, `LogstashPipeline`, `Script`, `Watch*`, `Index*` (retain `BuildDate` if still used by remaining code)
  - `ml.go` — delete all ML types (`Datafeed*`, `MLJob*`)
  - `transform.go` — delete all transform types
  - `enrich.go` — delete `EnrichPolicy`
  - Retain Kibana/Observability types (`action_connector.go`, `agent_builder.go`, `alert_rule.go`, `slo.go`) and truly custom types such as ingest processor structs in `ingest.go`

- **Update imports throughout**
  - Replace direct `github.com/elastic/go-elasticsearch/v8` imports with `github.com/elastic/go-elasticsearch/v8/typedapi/types` where appropriate
  - Remove unused `esapi` imports

## Capabilities

### New Capabilities
- `typed-client-final-cleanup`: Final infrastructure cleanup for the typed-client migration — removing the raw client accessor, obsolete helper functions, and redundant model types.

### Modified Capabilities
- (none — no Terraform resource or data source behavior changes; purely internal refactoring)

## Impact

- **Code**: `internal/clients/elasticsearch_scoped_client.go`, `internal/clients/elasticsearch/helpers.go`, `internal/models/*.go`, and all helper packages under `internal/clients/elasticsearch/`
- **APIs**: No Terraform resource or data source behavior changes
- **Dependencies**: Removes dependency on `esapi` response types from the raw client path
- **Tests**: Unit tests for `GetESClient` will need updating to expect a `*elasticsearch.TypedClient`; acceptance tests should be unaffected
