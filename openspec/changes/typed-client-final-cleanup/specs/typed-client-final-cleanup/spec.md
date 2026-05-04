## ADDED Requirements

### Requirement: GetESClient returns the typed client
`ElasticsearchScopedClient.GetESClient()` SHALL return `*elasticsearch.TypedClient` instead of `*elasticsearch.Client`.

#### Scenario: Typed client is returned
- **WHEN** `GetESClient()` is called on a configured `ElasticsearchScopedClient`
- **THEN** it returns a non-nil `*elasticsearch.TypedClient`

### Requirement: Raw client field removed from ElasticsearchScopedClient
`ElasticsearchScopedClient` SHALL no longer contain a raw `*elasticsearch.Client` field.

#### Scenario: Struct definition contains only typed client
- **WHEN** inspecting the `ElasticsearchScopedClient` struct definition
- **THEN** it does NOT contain a field of type `*elasticsearch.Client` from `github.com/elastic/go-elasticsearch/v8`

### Requirement: Obsolete helper functions deleted
`internal/clients/elasticsearch/helpers.go` SHALL be removed entirely, and no equivalent `doFWWrite` or `doSDKWrite` functions SHALL exist elsewhere in the codebase.

#### Scenario: Helpers file is absent
- **WHEN** checking for `internal/clients/elasticsearch/helpers.go`
- **THEN** the file does not exist

#### Scenario: Helper functions are not recreated
- **WHEN** searching the codebase for `doFWWrite` or `doSDKWrite`
- **THEN** no references exist outside of version control history

### Requirement: Redundant model types removed
The following model files SHALL be deleted, and their types SHALL not be referenced by any compiling code:
- `internal/models/ml.go`
- `internal/models/transform.go`
- `internal/models/enrich.go`

#### Scenario: ML model file is absent
- **WHEN** checking for `internal/models/ml.go`
- **THEN** the file does not exist

#### Scenario: Transform model file is absent
- **WHEN** checking for `internal/models/transform.go`
- **THEN** the file does not exist

#### Scenario: Enrich model file is absent
- **WHEN** checking for `internal/models/enrich.go`
- **THEN** the file does not exist

### Requirement: Unused types removed from models.go
`internal/models/models.go` SHALL retain only types that lack a `go-elasticsearch/v8/typedapi/types` equivalent or are documented as custom provider abstractions.

#### Scenario: Removed types have no remaining references
- **WHEN** searching the codebase for `models.ClusterInfo`, `models.User`, `models.Role`, `models.RoleMapping`, `models.APIKey`, `models.IndexTemplate`, `models.ComponentTemplate`, `models.Policy`, `models.SnapshotRepository`, `models.SnapshotPolicy`, `models.DataStream`, `models.LogstashPipeline`, `models.Script`, `models.Watch`
- **THEN** no references exist in compiling source files

#### Scenario: Custom types remain
- **WHEN** inspecting `internal/models/models.go`
- **THEN** it still contains types such as `BuildDate` if they are still used by remaining code

### Requirement: serverInfo uses typed API
`ElasticsearchScopedClient.serverInfo()` SHALL use the typed client's `Info().Do(ctx)` method and unmarshal into `*types.InfoResponse` from `go-elasticsearch/v8/typedapi/types`.

#### Scenario: serverInfo does not use raw esapi
- **WHEN** inspecting `serverInfo()` implementation
- **THEN** it does NOT call `esClient.Info()` or import `github.com/elastic/go-elasticsearch/v8/esapi`

### Requirement: Imports updated site-wide
All source files SHALL import `github.com/elastic/go-elasticsearch/v8/typedapi/types` where typed API types are used, and SHALL NOT import `github.com/elastic/go-elasticsearch/v8/esapi` unless required by Kibana or Fleet code paths outside Elasticsearch scope.

#### Scenario: No stale esapi imports in Elasticsearch helpers
- **WHEN** inspecting all `.go` files under `internal/clients/elasticsearch/`
- **THEN** none import `github.com/elastic/go-elasticsearch/v8/esapi`

### Requirement: Project builds successfully
`make build` SHALL complete without errors after all deletions and modifications.

#### Scenario: Clean build
- **WHEN** running `make build`
- **THEN** the command exits with status 0 and produces no compilation errors

### Requirement: Lint checks pass
`make check-lint` SHALL complete without new lint failures introduced by this change.

#### Scenario: Lint passes
- **WHEN** running `make check-lint`
- **THEN** the command exits with status 0
