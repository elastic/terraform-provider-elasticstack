## Why

The provider currently uses the raw `esapi` (untyped) client via `GetESClient()` for all ML API calls. The `go-elasticsearch/v8` Typed API (`elasticsearch.TypedClient`) provides strongly-typed request/response structs that eliminate hand-written JSON marshaling, reduce runtime errors, and enable IDE-driven refactoring. Phase 5 of the incremental typed-client migration targets the ML APIs.

## What Changes

- Migrate all functions in `internal/clients/elasticsearch/ml_job.go` from raw `esapi` to the typed client (`GetESTypedClient()`):
  - `OpenMLJob` → `typedapi.ML.OpenJob(...).Do(ctx)`
  - `PutDatafeed` → `typedapi.ML.PutDatafeed(...).Do(ctx)`
  - `CloseMLJob` → `typedapi.ML.CloseJob(...).Do(ctx)`
  - `GetMLJobStats` → `typedapi.ML.GetJobStats(...).Do(ctx)`
  - `GetDatafeed` → `typedapi.ML.GetDatafeeds(...).Do(ctx)`
  - `UpdateDatafeed` → `typedapi.ML.UpdateDatafeed(...).Do(ctx)`
  - `DeleteDatafeed` → `typedapi.ML.DeleteDatafeed(...).Do(ctx)`
  - `StopDatafeed` → `typedapi.ML.StopDatafeed(...).Do(ctx)`
  - `StartDatafeed` → `typedapi.ML.StartDatafeed(...).Do(ctx)`
  - `GetDatafeedStats` → `typedapi.ML.GetDatafeedStats(...).Do(ctx)`
- Migrate `internal/elasticsearch/ml/anomalydetectionjob/create.go`, `read.go`, `update.go`, and `delete.go` from direct `esClient.ML.*` raw calls to the typed client.
- Replace custom request/response model structs (`Datafeed`, `DatafeedCreateRequest`, `DatafeedUpdateRequest`, `DatafeedStats`, `DatafeedStatsResponse`, `MLJob`, `MLJobStats`) with typed API types (`types.Datafeed`, `types.Job`, etc.) where possible, or remove them entirely.
- Update downstream ML resource files (`datafeed`, `jobstate`, `datafeed_state`) that consume the migrated helpers so they continue to compile and behave identically.
- No Terraform user-facing schema or behavior changes.

## Capabilities

### New Capabilities
- `typed-client-ml`: Migrate all ML API calls and the ML anomaly detection job resource from the raw `esapi` client to the `go-elasticsearch` Typed API, and remove redundant custom ML model structs.

### Modified Capabilities
- _(none — no spec-level requirement changes; this is a pure implementation migration)_

## Impact

- **Code**: `internal/clients/elasticsearch/ml_job.go`, `internal/elasticsearch/ml/anomalydetectionjob/*.go`, `internal/elasticsearch/ml/datafeed/*.go`, `internal/elasticsearch/ml/jobstate/*.go`, `internal/elasticsearch/ml/datafeed_state/*.go`.
- **Models**: `internal/models/ml.go` — custom ML model structs become redundant and will be removed.
- **APIs**: No Terraform resource or data source behavior changes; all existing acceptance tests should continue to pass.
- **Dependencies**: Relies on the existing `typed-client-bootstrap` bridge (`GetESTypedClient()`).
