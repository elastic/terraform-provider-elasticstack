## ADDED Requirements

### Requirement: Transform helper functions use typed client
All functions in `internal/clients/elasticsearch/transform.go` SHALL use `GetESTypedClient()` and SHALL call the corresponding typed API methods instead of raw `esapi` methods.

#### Scenario: PutTransform uses typed client
- **WHEN** `PutTransform` is invoked
- **THEN** it calls `client.Transform.PutTransform(...).Do(ctx)` instead of `esClient.TransformPutTransform(...)`

#### Scenario: GetTransform uses typed client
- **WHEN** `GetTransform` is invoked
- **THEN** it calls `client.Transform.GetTransform(...).Do(ctx)` and returns a typed `*types.Transform` instead of decoding into a custom model

#### Scenario: GetTransformStats uses typed client
- **WHEN** `GetTransformStats` is invoked
- **THEN** it calls `client.Transform.GetTransformStats(...).Do(ctx)` and returns a typed `*types.TransformStats` instead of decoding into a custom model

#### Scenario: UpdateTransform uses typed client
- **WHEN** `UpdateTransform` is invoked
- **THEN** it calls `client.Transform.UpdateTransform(...).Do(ctx)` instead of `esClient.TransformUpdateTransform(...)`

#### Scenario: DeleteTransform uses typed client
- **WHEN** `DeleteTransform` is invoked
- **THEN** it calls `client.Transform.DeleteTransform(...).Force(true).Do(ctx)` instead of `esClient.TransformDeleteTransform(...)`

#### Scenario: startTransform uses typed client
- **WHEN** `startTransform` is invoked
- **THEN** it calls `client.Transform.StartTransform(...).Do(ctx)` instead of `esClient.TransformStartTransform(...)`

#### Scenario: stopTransform uses typed client
- **WHEN** `stopTransform` is invoked
- **THEN** it calls `client.Transform.StopTransform(...).Do(ctx)` instead of `esClient.TransformStopTransform(...)`

### Requirement: Transform resource uses typed client
`internal/elasticsearch/transform/transform.go` SHALL consume the migrated helpers and SHALL NOT call raw `esapi` transform methods directly.

#### Scenario: Create uses typed PutTransform
- **WHEN** the transform resource creates a transform
- **THEN** it calls `elasticsearch.PutTransform(...)` which uses the typed client internally

#### Scenario: Read uses typed GetTransform and GetTransformStats
- **WHEN** the transform resource reads a transform
- **THEN** it calls `elasticsearch.GetTransform(...)` and `elasticsearch.GetTransformStats(...)` which use the typed client internally

#### Scenario: Update uses typed UpdateTransform
- **WHEN** the transform resource updates a transform
- **THEN** it calls `elasticsearch.UpdateTransform(...)` which uses the typed client internally

#### Scenario: Delete uses typed DeleteTransform
- **WHEN** the transform resource deletes a transform
- **THEN** it calls `elasticsearch.DeleteTransform(...)` which uses the typed client internally

### Requirement: Redundant custom transform model structs removed
`internal/models/transform.go` SHALL be deleted, and the provider SHALL NOT contain custom structs `Transform`, `TransformSource`, `TransformDestination`, `TransformAlias`, `TransformRetentionPolicy`, `TransformRetentionPolicyTime`, `TransformSync`, `TransformSyncTime`, `TransformSettings`, `PutTransformParams`, `UpdateTransformParams`, `GetTransformResponse`, `TransformStats`, or `GetTransformStatsResponse`.

#### Scenario: Transform model file is absent
- **WHEN** checking for `internal/models/transform.go`
- **THEN** the file does not exist

#### Scenario: Custom transform types have no remaining references
- **WHEN** searching the codebase for `models.Transform`, `models.TransformStats`, or other custom transform types
- **THEN** no references exist in compiling source files

### Requirement: Transform acceptance tests compile and pass
`internal/elasticsearch/transform/transform_test.go` SHALL compile successfully after the migration, and the transform acceptance test suite SHALL pass.

#### Scenario: CheckDestroy uses typed client
- **WHEN** `checkResourceTransformDestroy` runs
- **THEN** it uses the typed client to verify transform deletion

#### Scenario: Acceptance tests pass after migration
- **WHEN** running the transform acceptance tests
- **THEN** all tests pass without errors

### Requirement: Project builds successfully after transform migration
`make build` SHALL complete without errors after all transform files are migrated and redundant models are removed.

#### Scenario: Clean build after transform migration
- **WHEN** running `make build`
- **THEN** the command exits with status 0 and produces no compilation errors

### Requirement: Lint checks pass after transform migration
`make check-lint` SHALL complete without new lint failures introduced by the typed-client migration.

#### Scenario: Lint passes after transform migration
- **WHEN** running `make check-lint`
- **THEN** the command exits with status 0
