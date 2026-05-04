## ADDED Requirements

### Requirement: Cluster info helper uses typed client
`GetClusterInfo` SHALL use the typed client (`typedapi.Core.Info().Do(ctx)`) to retrieve cluster metadata. It SHALL return `*info.Response` from the typed client. When the API returns an error, the helper SHALL surface the error in Terraform diagnostics.

#### Scenario: Successful cluster info retrieval
- **WHEN** `GetClusterInfo` is called with a valid `*clients.ElasticsearchScopedClient`
- **THEN** it returns a non-nil `*info.Response` and no error diagnostics

#### Scenario: API failure on cluster info
- **GIVEN** the Elasticsearch cluster info API returns an error
- **WHEN** `GetClusterInfo` processes the response
- **THEN** it returns nil and Terraform diagnostics include the API error

### Requirement: Snapshot repository helpers use typed client
`PutSnapshotRepository`, `GetSnapshotRepository`, and `DeleteSnapshotRepository` SHALL use the typed snapshot repository APIs (`Snapshot.CreateRepository`, `Snapshot.GetRepository`, `Snapshot.DeleteRepository`). `GetSnapshotRepository` SHALL return the repository's typed response and SHALL return `nil` with no error when the repository is not found.

#### Scenario: Create or update snapshot repository
- **WHEN** `PutSnapshotRepository` is called with a valid repository configuration
- **THEN** it calls the typed `Snapshot.CreateRepository` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing snapshot repository
- **WHEN** `GetSnapshotRepository` is called for an existing repository
- **THEN** it returns the repository data and no error diagnostics

#### Scenario: Read missing snapshot repository
- **GIVEN** the requested snapshot repository does not exist
- **WHEN** `GetSnapshotRepository` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete snapshot repository
- **WHEN** `DeleteSnapshotRepository` is called with a valid repository name
- **THEN** it calls the typed `Snapshot.DeleteRepository` API
- **AND** returns no error diagnostics on success

### Requirement: SLM helpers use typed client
`PutSlm`, `GetSlm`, and `DeleteSlm` SHALL use the typed SLM APIs (`Slm.PutLifecycle`, `Slm.GetLifecycle`, `Slm.DeleteLifecycle`). `GetSlm` SHALL return the policy's typed response and SHALL return `nil` with no error when the policy is not found.

#### Scenario: Create or update SLM policy
- **WHEN** `PutSlm` is called with a valid SLM policy
- **THEN** it calls the typed `Slm.PutLifecycle` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing SLM policy
- **WHEN** `GetSlm` is called for an existing policy
- **THEN** it returns the policy data and no error diagnostics

#### Scenario: Read missing SLM policy
- **GIVEN** the requested SLM policy does not exist
- **WHEN** `GetSlm` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete SLM policy
- **WHEN** `DeleteSlm` is called with a valid policy name
- **THEN** it calls the typed `Slm.DeleteLifecycle` API
- **AND** returns no error diagnostics on success

### Requirement: Cluster settings helpers use typed client
`PutSettings` and `GetSettings` SHALL use the typed cluster settings APIs (`Cluster.PutSettings`, `Cluster.GetSettings`). `GetSettings` SHALL request flat settings (`flat_settings=true`) and return them as a `map[string]any`.

#### Scenario: Update cluster settings
- **WHEN** `PutSettings` is called with a settings map
- **THEN** it calls the typed `Cluster.PutSettings` API
- **AND** returns no error diagnostics on success

#### Scenario: Read cluster settings with flat settings
- **WHEN** `GetSettings` is called
- **THEN** it calls the typed `Cluster.GetSettings` API with `flat_settings=true`
- **AND** returns the settings as `map[string]any`

### Requirement: Script helpers use typed client
`GetScript`, `PutScript`, and `DeleteScript` SHALL use the typed script APIs (`Core.GetScript`, `Core.PutScript`, `Core.DeleteScript`). `GetScript` SHALL return `*types.StoredScript` and SHALL return `nil` with no error when the script is not found. `PutScript` SHALL build the request body from `*types.StoredScript`.

#### Scenario: Create or update stored script
- **WHEN** `PutScript` is called with a valid `*types.StoredScript`
- **THEN** it calls the typed `Core.PutScript` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing stored script
- **WHEN** `GetScript` is called for an existing script
- **THEN** it returns `*types.StoredScript` and no error diagnostics

#### Scenario: Read missing stored script
- **GIVEN** the requested stored script does not exist
- **WHEN** `GetScript` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete stored script
- **WHEN** `DeleteScript` is called with a valid script ID
- **THEN** it calls the typed `Core.DeleteScript` API
- **AND** returns no error diagnostics on success

### Requirement: Custom cluster model types are removed
The custom model types `ClusterInfo`, `SnapshotRepository`, `SnapshotPolicy`, and `Script` in `internal/models/models.go` SHALL be removed once all callers have been migrated to typed client equivalents. No remaining code SHALL reference these types after the migration.

#### Scenario: Build succeeds after model removal
- **GIVEN** all callers have been updated to use typed client types
- **WHEN** the custom models are removed from `internal/models/models.go`
- **THEN** `make build` completes successfully
