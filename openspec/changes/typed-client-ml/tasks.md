## 1. Migrate `internal/clients/elasticsearch/ml_job.go` helpers

- [x] 1.1 Replace `esClient.ML.OpenJob` with `typedapi.ML.OpenJob(...).Do(ctx)` in `OpenMLJob`
- [x] 1.2 Replace `esClient.ML.PutDatafeed` with `typedapi.ML.PutDatafeed(...).Do(ctx)` in `PutDatafeed`; accept typed `types.DatafeedConfig` instead of `models.DatafeedCreateRequest`
- [x] 1.3 Replace `esClient.ML.CloseJob` with `typedapi.ML.CloseJob(...).Do(ctx)` in `CloseMLJob`
- [x] 1.4 Replace `esClient.ML.GetJobStats` with `typedapi.ML.GetJobStats(...).Do(ctx)` in `GetMLJobStats`; return typed `types.JobStats` instead of `models.MLJob`
- [x] 1.5 Replace `esClient.ML.GetDatafeeds` with `typedapi.ML.GetDatafeeds(...).Do(ctx)` in `GetDatafeed`; return typed `types.Datafeed` instead of `models.Datafeed`
- [x] 1.6 Replace `esClient.ML.UpdateDatafeed` with `typedapi.ML.UpdateDatafeed(...).Do(ctx)` in `UpdateDatafeed`; accept typed `types.DatafeedConfig` instead of `models.DatafeedUpdateRequest`
- [x] 1.7 Replace `esClient.ML.DeleteDatafeed` with `typedapi.ML.DeleteDatafeed(...).Do(ctx)` in `DeleteDatafeed`
- [x] 1.8 Replace `esClient.ML.StopDatafeed` with `typedapi.ML.StopDatafeed(...).Do(ctx)` in `StopDatafeed`
- [x] 1.9 Replace `esClient.ML.StartDatafeed` with `typedapi.ML.StartDatafeed(...).Do(ctx)` in `StartDatafeed`
- [x] 1.10 Replace `esClient.ML.GetDatafeedStats` with `typedapi.ML.GetDatafeedStats(...).Do(ctx)` in `GetDatafeedStats`; return typed `types.DatafeedStats` instead of `models.DatafeedStats`
- [x] 1.11 Update all helper signatures to accept `*clients.ElasticsearchScopedClient` and call `GetESTypedClient()` instead of `GetESClient()`
- [x] 1.12 Remove unused `bytes`, `encoding/json`, `net/http`, and `esapi` imports from `ml_job.go` where possible

## 2. Migrate `internal/elasticsearch/ml/anomalydetectionjob` resource files

- [x] 2.1 Replace raw `esClient.ML.PutJob` with typed client in `create.go`; build request with `types.JobConfig` directly
- [x] 2.2 Replace raw `esClient.ML.GetJobs` with typed client in `read.go`; decode into `types.Job` instead of custom `APIModel`
- [x] 2.3 Replace raw `esClient.ML.UpdateJob` with typed client in `update.go`; build update body with typed struct
- [x] 2.4 Replace raw `esClient.ML.CloseJob` and `esClient.ML.DeleteJob` with typed client in `delete.go`
- [x] 2.5 Remove raw `esapi` import and JSON marshal/unmarshal boilerplate from `anomalydetectionjob/*.go`
- [x] 2.6 Ensure `models_api.go` and `models_tf.go` (if any) are updated or removed in favor of typed API structs

## 3. Update downstream ML resource files that consume helpers

- [x] 3.1 Update `internal/elasticsearch/ml/datafeed/create.go` to pass typed request to `elasticsearch.PutDatafeed`
- [x] 3.2 Update `internal/elasticsearch/ml/datafeed/read.go` to consume typed response from `elasticsearch.GetDatafeed`
- [x] 3.3 Update `internal/elasticsearch/ml/datafeed/update.go` to pass typed request to `elasticsearch.UpdateDatafeed` and `elasticsearch.StartDatafeed`
- [x] 3.4 Update `internal/elasticsearch/ml/datafeed/delete.go` to consume typed helpers for stop and delete
- [x] 3.5 Update `internal/elasticsearch/ml/datafeed/state_utils.go` to consume typed response from `elasticsearch.GetDatafeedStats`
- [x] 3.6 Update `internal/elasticsearch/ml/jobstate/update.go` to consume typed helpers for open and close
- [x] 3.7 Update `internal/elasticsearch/ml/jobstate/state_utils.go` to consume typed response from `elasticsearch.GetMLJobStats`
- [x] 3.8 Update `internal/elasticsearch/ml/datafeed_state/*.go` to consume typed helpers for stats, start, and stop

## 4. Clean up redundant custom models

- [x] 4.1 Remove `Datafeed`, `DatafeedCreateRequest`, `DatafeedUpdateRequest`, `DatafeedStats`, `DatafeedStatsResponse`, `MLJob`, and `MLJobStats` from `internal/models/ml.go`
- [x] 4.2 Verify no other packages import the removed model structs
- [x] 4.3 Run `go mod tidy` and `make build` to confirm everything compiles

## 5. Verify behavior and tests

- [x] 5.1 Run `make check-lint` and fix any issues
- [x] 5.2 Run `make build` to ensure no compilation errors
- [x] 5.3 Run ML unit tests: `go test ./internal/elasticsearch/ml/...`
- [x] 5.4 Run ML acceptance tests against a live Elasticsearch cluster
- [x] 5.5 Verify error handling for 404/not-found scenarios remains unchanged
- [x] 5.6 Verify force flags, timeouts, and optional parameters are still passed correctly

## 6. Post-review bug fixes

- [x] 6.1 Fix timeout string format: use `ms` suffix instead of `Go's native format` in `CloseMLJob`, `StopDatafeed`, `StartDatafeed`
- [x] 6.2 Fix `fromTypedJob` missing `Groups` field mapping
- [x] 6.3 Fix datafeed query normalization: use raw JSON from API to avoid term shorthand expansion
- [x] 6.4 Fix `categorization_examples_limit` dropped on update: send update as raw JSON to include field missing from `types.AnalysisMemoryLimit`
- [x] 6.5 Add `AllowNoMatch(true)` to `GetJobs` call in `read.go`
- [x] 6.6 Add error guard after Groups `ElementsAs` in `BuildFromPlan`
- [x] 6.7 Add 404 guard to `DeleteDatafeed`
- [x] 6.8 Fix typo in update warning message: "No changed detected to updateble fields" → "No changes detected to updatable fields"
