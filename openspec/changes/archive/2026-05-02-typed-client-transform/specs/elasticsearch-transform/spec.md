## ADDED Requirements

### Requirement: Transform create uses typed client
`PutTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.PutTransform`). It SHALL pass the transform request body via the typed API's `.Raw()` method so that all fields — including `destination.aliases` — are preserved. Query parameters (`defer_validation`, `timeout`) SHALL be set via the typed API builder methods. The helper SHALL surface API errors as Terraform diagnostics.

#### Scenario: Typed API create with aliases
- **GIVEN** a transform configuration that includes `destination.aliases`
- **WHEN** `PutTransform` is called
- **THEN** it calls the typed `Transform.PutTransform` API
- **AND** the request body includes the `aliases` field
- **AND** it returns no error diagnostics on success

#### Scenario: Typed API create error handling
- **GIVEN** the Put Transform API returns an error
- **WHEN** `PutTransform` processes the response
- **THEN** it returns Terraform diagnostics containing the API error

### Requirement: Transform read uses typed client
`GetTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.GetTransform`) via `.Perform()` to obtain the raw HTTP response. It SHALL decode the response body into the existing `models.GetTransformResponse` structure so that all fields — including `destination.aliases` — are read correctly. When the API returns HTTP 404, the helper SHALL return `nil` with no error diagnostics.

#### Scenario: Typed API read existing transform
- **GIVEN** an existing transform with `destination.aliases`
- **WHEN** `GetTransform` is called
- **THEN** it calls the typed `Transform.GetTransform` API
- **AND** the returned transform includes the `aliases` field
- **AND** it returns no error diagnostics

#### Scenario: Typed API read missing transform
- **GIVEN** the requested transform does not exist
- **WHEN** `GetTransform` is called
- **THEN** it returns `nil` and no error diagnostics

### Requirement: Transform stats uses typed client
`GetTransformStats` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.GetTransformStats`). It SHALL search the returned `[]types.TransformStats` for the matching transform ID and derive the `enabled` state from the `state` field ("started" or "indexing" means enabled).

#### Scenario: Typed API stats for started transform
- **GIVEN** a transform whose state is "started"
- **WHEN** `GetTransformStats` is called
- **THEN** it calls the typed `Transform.GetTransformStats` API
- **AND** it returns stats with `IsStarted() == true`

#### Scenario: Typed API stats for stopped transform
- **GIVEN** a transform whose state is "stopped"
- **WHEN** `GetTransformStats` is called
- **THEN** it calls the typed `Transform.GetTransformStats` API
- **AND** it returns stats with `IsStarted() == false`

### Requirement: Transform update uses typed client
`UpdateTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.UpdateTransform`). It SHALL pass the transform request body via the typed API's `.Raw()` method so that all updatable fields are preserved. Query parameters (`defer_validation`, `timeout`) SHALL be set via the typed API builder methods. After a successful update, it SHALL optionally call `startTransform` or `stopTransform` based on the `enabled` change exactly as today.

#### Scenario: Typed API update with enabled change
- **GIVEN** an existing transform and `enabled` changed to `false`
- **WHEN** `UpdateTransform` is called
- **THEN** it calls the typed `Transform.UpdateTransform` API
- **AND** it calls `stopTransform` after the update succeeds
- **AND** it returns no error diagnostics on success

#### Scenario: Typed API update without enabled change
- **GIVEN** an existing transform and `enabled` is unchanged
- **WHEN** `UpdateTransform` is called
- **THEN** it calls the typed `Transform.UpdateTransform` API
- **AND** it does NOT call `startTransform` or `stopTransform`

### Requirement: Transform delete uses typed client
`DeleteTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.DeleteTransform`). It SHALL pass `force=true` via the typed API builder method. The helper SHALL surface API errors as Terraform diagnostics.

#### Scenario: Typed API delete transform
- **GIVEN** an existing transform
- **WHEN** `DeleteTransform` is called
- **THEN** it calls the typed `Transform.DeleteTransform` API with `force=true`
- **AND** it returns no error diagnostics on success

#### Scenario: Typed API delete error handling
- **GIVEN** the Delete Transform API returns an error
- **WHEN** `DeleteTransform` processes the response
- **THEN** it returns Terraform diagnostics containing the API error

### Requirement: Transform start uses typed client
`startTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.StartTransform`). It SHALL pass the `timeout` parameter via the typed API builder method. The helper SHALL surface API errors as Terraform diagnostics.

#### Scenario: Typed API start transform
- **GIVEN** a stopped transform
- **WHEN** `startTransform` is called
- **THEN** it calls the typed `Transform.StartTransform` API
- **AND** it returns no error diagnostics on success

### Requirement: Transform stop uses typed client
`stopTransform` SHALL use the go-elasticsearch Typed API (`elasticsearch.TypedClient.Transform.StopTransform`). It SHALL pass the `timeout` parameter via the typed API builder method. The helper SHALL surface API errors as Terraform diagnostics.

#### Scenario: Typed API stop transform
- **GIVEN** a started transform
- **WHEN** `stopTransform` is called
- **THEN** it calls the typed `Transform.StopTransform` API
- **AND** it returns no error diagnostics on success

### Requirement: Unused transform model types are removed
The custom model types `models.PutTransformParams`, `models.UpdateTransformParams`, `models.TransformStats`, and `models.GetTransformStatsResponse` SHALL be removed once all callers have been migrated to typed client equivalents or to inline parameters. `models.Transform` and `models.GetTransformResponse` MAY be retained for JSON body construction and response decoding where the typed API types do not fully cover provider-supported fields.

#### Scenario: Build succeeds after model cleanup
- **GIVEN** all callers have been updated to use typed client types or inline params
- **WHEN** the unused custom models are removed from `internal/models/transform.go`
- **THEN** `make build` completes successfully
