# typed-client-easy-wins Specification

## Purpose
TBD - created by archiving change typed-client-easy-wins. Update Purpose after archive.
## Requirements
### Requirement: Inference helpers use typed client
The `inference.go` helper functions SHALL use the typed Elasticsearch client (`GetESTypedClient()`) for all inference endpoint API operations.

#### Scenario: PutInferenceEndpoint uses typed API
- **WHEN** `PutInferenceEndpoint` creates or updates an inference endpoint
- **THEN** it SHALL call `typedapi.Inference.Put()` with `types.InferenceEndpoint`
- **AND** it SHALL NOT use `esapi.InferencePut` or manual JSON marshaling

#### Scenario: GetInferenceEndpoint uses typed API
- **WHEN** `GetInferenceEndpoint` reads an inference endpoint
- **THEN** it SHALL call `typedapi.Inference.Get()` and use `types.InferenceEndpointInfo` for the response
- **AND** it SHALL NOT decode JSON from an `*esapi.Response`

#### Scenario: UpdateInferenceEndpoint uses typed API
- **WHEN** `UpdateInferenceEndpoint` updates an inference endpoint
- **THEN** it SHALL call `typedapi.Inference.Update()`
- **AND** it SHALL NOT use `esapi.InferenceUpdate`
- **AND** it MAY use `.Raw()` to supply a custom JSON body when `types.InferenceEndpoint` serializes `service` (an immutable field the update API rejects); such usage SHALL be documented inline with a design rationale

#### Scenario: DeleteInferenceEndpoint uses typed API
- **WHEN** `DeleteInferenceEndpoint` deletes an inference endpoint
- **THEN** it SHALL call `typedapi.Inference.Delete()`
- **AND** it SHALL NOT use `esapi.InferenceDelete`

### Requirement: Logstash helpers use typed client
The `logstash.go` helper functions SHALL use the typed Elasticsearch client for all logstash pipeline API operations.

#### Scenario: PutLogstashPipeline uses typed API
- **WHEN** `PutLogstashPipeline` creates or updates a logstash pipeline
- **THEN** it SHALL call `typedapi.Logstash.PutPipeline()` with `types.LogstashPipeline`
- **AND** it SHALL NOT use `esapi.LogstashPutPipeline`
- **AND** it MAY use `.Raw()` to preserve pipeline settings that `types.LogstashPipeline` does not fully support; such usage SHALL be documented inline with a design rationale

#### Scenario: GetLogstashPipeline uses typed API
- **WHEN** `GetLogstashPipeline` reads a logstash pipeline
- **THEN** it SHALL call `typedapi.Logstash.GetPipeline()` and extract pipeline data from the response
- **AND** it SHALL NOT decode JSON from an `*esapi.Response`
- **AND** it MAY decode from the raw `*http.Response` provided by `.Perform()` when `types.LogstashPipeline` does not fully support all pipeline settings; such usage SHALL be documented inline with a design rationale

#### Scenario: DeleteLogstashPipeline uses typed API
- **WHEN** `DeleteLogstashPipeline` deletes a logstash pipeline
- **THEN** it SHALL call `typedapi.Logstash.DeletePipeline()`
- **AND** it SHALL NOT use `esapi.LogstashDeletePipeline`

### Requirement: Enrich helpers use typed client
The `enrich.go` helper functions SHALL use the typed Elasticsearch client for all enrich policy API operations.

#### Scenario: GetEnrichPolicy uses typed API
- **WHEN** `GetEnrichPolicy` reads an enrich policy
- **THEN** it SHALL call `typedapi.Enrich.GetPolicy()` and use `types.EnrichPolicy` for the response
- **AND** it SHALL NOT decode JSON from an `*esapi.Response`

#### Scenario: PutEnrichPolicy uses typed API
- **WHEN** `PutEnrichPolicy` creates an enrich policy
- **THEN** it SHALL call `typedapi.Enrich.PutPolicy()` with `types.EnrichPolicy`
- **AND** it SHALL NOT use `esapi.EnrichPutPolicy` or manual JSON marshaling

#### Scenario: DeleteEnrichPolicy uses typed API
- **WHEN** `DeleteEnrichPolicy` deletes an enrich policy
- **THEN** it SHALL call `typedapi.Enrich.DeletePolicy()`
- **AND** it SHALL NOT use `esapi.EnrichDeletePolicy`

#### Scenario: ExecuteEnrichPolicy uses typed API
- **WHEN** `ExecuteEnrichPolicy` executes an enrich policy
- **THEN** it SHALL call `typedapi.Enrich.ExecutePolicy()` with `wait_for_completion=true`
- **AND** it SHALL NOT use `esapi.EnrichExecutePolicy`

### Requirement: Watch helpers use typed client
The `watch.go` helper functions SHALL use the typed Elasticsearch client for all watcher API operations.

#### Scenario: PutWatch uses typed API
- **WHEN** `PutWatch` creates or updates a watch
- **THEN** it SHALL call `typedapi.Watcher.PutWatch()` with `types.Watch` or equivalent typed request fields
- **AND** it SHALL NOT use `esapi.Watcher.PutWatch`
- **AND** it MAY use `.Raw()` to preserve dynamic watch body fields (trigger/input/condition/actions/metadata/transform) that the Terraform schema stores as normalized JSON; such usage SHALL be documented inline with a design rationale

#### Scenario: GetWatch uses typed API
- **WHEN** `GetWatch` reads a watch
- **THEN** it SHALL call `typedapi.Watcher.GetWatch()` and use `types.Watch` for the response
- **AND** it SHALL NOT decode JSON from an `*esapi.Response`

#### Scenario: DeleteWatch uses typed API
- **WHEN** `DeleteWatch` deletes a watch
- **THEN** it SHALL call `typedapi.Watcher.DeleteWatch()`
- **AND** it SHALL NOT use `esapi.Watcher.DeleteWatch`
- **AND** a 404 response SHALL be treated as success (no error diagnostic)

### Requirement: Acceptance-test compatibility
Where typed-client migration changes inference endpoint behavior (e.g. update body shape) or where acceptance tests depend on real third-party API credentials, the test suite SHALL be updated to skip affected update steps when only a fake API key is available.

#### Scenario: Inference acceptance tests skip update steps with fake API key
- **WHEN** the acceptance test suite runs with the default fake OpenAI API key
- **THEN** update/plan steps that would trigger real service validation SHALL be skipped via `SkipFunc`
- **AND** all remaining test steps SHALL still execute correctly

### Requirement: Preserved error semantics
All migrated helpers SHALL preserve the existing error handling and not-found semantics observed by downstream Terraform resources.

#### Scenario: Get helpers return nil on not found
- **WHEN** a Get helper queries a resource that does not exist
- **THEN** it SHALL return `nil` for the resource and no error diagnostic

#### Scenario: Delete helpers succeed on not found
- **WHEN** a Delete helper targets a resource that does not exist
- **THEN** it SHALL return empty diagnostics without error

### Requirement: No raw esapi usage in migrated files
After migration, the four migrated helper files SHALL contain no calls to `GetESClient()` and no usage of `esapi` types for the APIs they cover.

#### Scenario: inference.go contains no raw esapi calls
- **GIVEN** `internal/clients/elasticsearch/inference.go` has been migrated
- **WHEN** inspecting the file for `GetESClient` and `esapi.Inference` usage
- **THEN** no such usage SHALL remain

#### Scenario: logstash.go contains no raw esapi calls
- **GIVEN** `internal/clients/elasticsearch/logstash.go` has been migrated
- **WHEN** inspecting the file for `GetESClient` and `esapi.Logstash` usage
- **THEN** no such usage SHALL remain

#### Scenario: enrich.go contains no raw esapi calls
- **GIVEN** `internal/clients/elasticsearch/enrich.go` has been migrated
- **WHEN** inspecting the file for `GetESClient` and `esapi.Enrich` usage
- **THEN** no such usage SHALL remain

#### Scenario: watch.go contains no raw esapi calls
- **GIVEN** `internal/clients/elasticsearch/watch.go` has been migrated
- **WHEN** inspecting the file for `GetESClient` and `esapi.Watcher` usage
- **THEN** no such usage SHALL remain

