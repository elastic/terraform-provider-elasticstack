## ADDED Requirements

### Requirement: ML helper functions use typed client
All functions in `internal/clients/elasticsearch/ml_job.go` SHALL use `GetESTypedClient()` and SHALL call the corresponding typed API methods instead of raw `esapi` methods.

#### Scenario: OpenMLJob uses typed client
- **WHEN** `OpenMLJob` is invoked
- **THEN** it calls `client.ML.OpenJob(...).Do(ctx)` instead of `esClient.ML.OpenJob(...)`

#### Scenario: PutDatafeed uses typed client
- **WHEN** `PutDatafeed` is invoked
- **THEN** it calls `client.ML.PutDatafeed(...).Raw(body).Do(ctx)` instead of `esClient.ML.PutDatafeed(...)`, accepting a typed `DatafeedRequest` (client-layer struct) marshalled to raw bytes so that `Query` is preserved without normalisation through `types.Query`

#### Scenario: CloseMLJob uses typed client
- **WHEN** `CloseMLJob` is invoked
- **THEN** it calls `client.ML.CloseJob(...).Do(ctx)` instead of `esClient.ML.CloseJob(...)`

#### Scenario: GetMLJobStats uses typed client
- **WHEN** `GetMLJobStats` is invoked
- **THEN** it calls `client.ML.GetJobStats(...).Do(ctx)` and returns a typed `*types.JobStats` instead of decoding into a custom model

#### Scenario: GetDatafeed uses typed client
- **WHEN** `GetDatafeed` is invoked
- **THEN** it calls `client.ML.GetDatafeeds(...).Perform(ctx)` (raw response) and returns `*MLDatafeedResponse` — a thin wrapper around `*types.MLDatafeed` that also carries `QueryRaw json.RawMessage` to preserve the query exactly as returned by Elasticsearch without re-normalisation

#### Scenario: UpdateDatafeed uses typed client
- **WHEN** `UpdateDatafeed` is invoked
- **THEN** it calls `client.ML.UpdateDatafeed(...).Raw(body).Do(ctx)` instead of `esClient.ML.UpdateDatafeed(...)`, accepting a typed `DatafeedRequest` (client-layer struct, `JobID` cleared) marshalled to raw bytes for the same query-preservation reason as `PutDatafeed`

#### Scenario: DeleteDatafeed uses typed client
- **WHEN** `DeleteDatafeed` is invoked
- **THEN** it calls `client.ML.DeleteDatafeed(...).Do(ctx)` instead of `esClient.ML.DeleteDatafeed(...)`

#### Scenario: StopDatafeed uses typed client
- **WHEN** `StopDatafeed` is invoked
- **THEN** it calls `client.ML.StopDatafeed(...).Do(ctx)` instead of `esClient.ML.StopDatafeed(...)`

#### Scenario: StartDatafeed uses typed client
- **WHEN** `StartDatafeed` is invoked
- **THEN** it calls `client.ML.StartDatafeed(...).Do(ctx)` instead of `esClient.ML.StartDatafeed(...)`

#### Scenario: GetDatafeedStats uses typed client
- **WHEN** `GetDatafeedStats` is invoked
- **THEN** it calls `client.ML.GetDatafeedStats(...).Do(ctx)` and returns a typed `*types.DatafeedStats` instead of decoding into a custom model

### Requirement: Anomaly detection job resource uses typed client
`internal/elasticsearch/ml/anomalydetectionjob/create.go`, `read.go`, `update.go`, and `delete.go` SHALL use `GetESTypedClient()` and SHALL NOT call raw `esapi` ML methods directly.

#### Scenario: Create uses typed PutJob
- **WHEN** the anomaly detection job resource creates a job
- **THEN** it builds a `types.JobConfig` request and calls `client.ML.PutJob(...).Do(ctx)` instead of `esClient.ML.PutJob(...)`

#### Scenario: Read uses typed GetJobs
- **WHEN** the anomaly detection job resource reads a job
- **THEN** it calls `client.ML.GetJobs(...).Do(ctx)` and uses typed `types.Job` responses instead of manual JSON decoding into custom structs

#### Scenario: Update uses typed UpdateJob
- **WHEN** the anomaly detection job resource updates a job
- **THEN** it calls `client.ML.UpdateJob(...).Do(ctx)` instead of `esClient.ML.UpdateJob(...)`

#### Scenario: Delete uses typed CloseJob and DeleteJob
- **WHEN** the anomaly detection job resource deletes a job
- **THEN** it calls `client.ML.CloseJob(...).Do(ctx)` and `client.ML.DeleteJob(...).Do(ctx)` instead of raw `esClient.ML.CloseJob(...)` and `esClient.ML.DeleteJob(...)`

### Requirement: Redundant custom ML model structs removed
`internal/models/ml.go` SHALL be deleted, and the provider SHALL NOT contain custom structs `Datafeed`, `DatafeedCreateRequest`, `DatafeedUpdateRequest`, `DatafeedStats`, `DatafeedStatsResponse`, `MLJob`, or `MLJobStats`.

#### Scenario: ML model file is absent
- **WHEN** checking for `internal/models/ml.go`
- **THEN** the file does not exist

#### Scenario: Custom ML types have no remaining references
- **WHEN** searching the codebase for `models.Datafeed`, `models.DatafeedCreateRequest`, `models.DatafeedUpdateRequest`, `models.DatafeedStats`, `models.DatafeedStatsResponse`, `models.MLJob`, or `models.MLJobStats`
- **THEN** no references exist in compiling source files

### Requirement: Downstream ML resources compile against migrated helpers
`internal/elasticsearch/ml/datafeed`, `internal/elasticsearch/ml/jobstate`, and `internal/elasticsearch/ml/datafeed_state` SHALL compile successfully after `ml_job.go` helper signatures are updated.

#### Scenario: Datafeed resource compiles
- **WHEN** compiling the datafeed resource package
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Job state resource compiles
- **WHEN** compiling the jobstate resource package
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Datafeed state resource compiles
- **WHEN** compiling the datafeed_state resource package
- **THEN** it succeeds with no errors referencing the updated helper signatures

### Requirement: Project builds successfully after ML migration
`make build` SHALL complete without errors after all ML files are migrated and redundant models are removed.

#### Scenario: Clean build after ML migration
- **WHEN** running `make build`
- **THEN** the command exits with status 0 and produces no compilation errors

### Requirement: Lint checks pass after ML migration
`make check-lint` SHALL complete without new lint failures introduced by the typed-client migration.

#### Scenario: Lint passes after ML migration
- **WHEN** running `make check-lint`
- **THEN** the command exits with status 0
