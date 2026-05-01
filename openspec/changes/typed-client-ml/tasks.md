## 1. Migrate `internal/clients/elasticsearch/ml_job.go` helpers

- [ ] 1.1 Replace `esClient.ML.OpenJob` with `typedapi.ML.OpenJob(...).Do(ctx)` in `OpenMLJob`
- [ ] 1.2 Replace `esClient.ML.PutDatafeed` with `typedapi.ML.PutDatafeed(...).Do(ctx)` in `PutDatafeed`; accept typed `types.DatafeedConfig` instead of `models.DatafeedCreateRequest`
- [ ] 1.3 Replace `esClient.ML.CloseJob` with `typedapi.ML.CloseJob(...).Do(ctx)` in `CloseMLJob`
- [ ] 1.4 Replace `esClient.ML.GetJobStats` with `typedapi.ML.GetJobStats(...).Do(ctx)` in `GetMLJobStats`; return typed `types.JobStats` instead of `models.MLJob`
- [ ] 1.5 Replace `esClient.ML.GetDatafeeds` with `typedapi.ML.GetDatafeeds(...).Do(ctx)` in `GetDatafeed`; return typed `types.Datafeed` instead of `models.Datafeed`
- [ ] 1.6 Replace `esClient.ML.UpdateDatafeed` with `typedapi.ML.UpdateDatafeed(...).Do(ctx)` in `UpdateDatafeed`; accept typed `types.DatafeedConfig` instead of `models.DatafeedUpdateRequest`
- [ ] 1.7 Replace `esClient.ML.DeleteDatafeed` with `typedapi.ML.DeleteDatafeed(...).Do(ctx)` in `DeleteDatafeed`
- [ ] 1.8 Replace `esClient.ML.StopDatafeed` with `typedapi.ML.StopDatafeed(...).Do(ctx)` in `StopDatafeed`
- [ ] 1.9 Replace `esClient.ML.StartDatafeed` with `typedapi.ML.StartDatafeed(...).Do(ctx)` in `StartDatafeed`
- [ ] 1.10 Replace `esClient.ML.GetDatafeedStats` with `typedapi.ML.GetDatafeedStats(...).Do(ctx)` in `GetDatafeedStats`; return typed `types.DatafeedStats` instead of `models.DatafeedStats`
- [ ] 1.11 Update all helper signatures to accept `*clients.ElasticsearchScopedClient` and call `GetESTypedClient()` instead of `GetESClient()`
- [ ] 1.12 Remove unused `bytes`, `encoding/json`, `net/http`, and `esapi` imports from `ml_job.go` where possible

## 2. Migrate `internal/elasticsearch/ml/anomalydetectionjob` resource files

- [ ] 2.1 Replace raw `esClient.ML.PutJob` with typed client in `create.go`; build request with `types.JobConfig` directly
- [ ] 2.2 Replace raw `esClient.ML.GetJobs` with typed client in `read.go`; decode into `types.Job` instead of custom `APIModel`
- [ ] 2.3 Replace raw `esClient.ML.UpdateJob` with typed client in `update.go`; build update body with typed struct
- [ ] 2.4 Replace raw `esClient.ML.CloseJob` and `esClient.ML.DeleteJob` with typed client in `delete.go`
- [ ] 2.5 Remove raw `esapi` import and JSON marshal/unmarshal boilerplate from `anomalydetectionjob/*.go`
- [ ] 2.6 Ensure `models_api.go` and `models_tf.go` (if any) are updated or removed in favor of typed API structs

## 3. Update downstream ML resource files that consume helpers

- [ ] 3.1 Update `internal/elasticsearch/ml/datafeed/create.go` to pass typed request to `elasticsearch.PutDatafeed`
- [ ] 3.2 Update `internal/elasticsearch/ml/datafeed/read.go` to consume typed response from `elasticsearch.GetDatafeed`
- [ ] 3.3 Update `internal/elasticsearch/ml/datafeed/update.go` to pass typed request to `elasticsearch.UpdateDatafeed` and `elasticsearch.StartDatafeed`
- [ ] 3.4 Update `internal/elasticsearch/ml/datafeed/delete.go` to consume typed helpers for stop and delete
- [ ] 3.5 Update `internal/elasticsearch/ml/datafeed/state_utils.go` to consume typed response from `elasticsearch.GetDatafeedStats`
- [ ] 3.6 Update `internal/elasticsearch/ml/jobstate/update.go` to consume typed helpers for open and close
- [ ] 3.7 Update `internal/elasticsearch/ml/jobstate/state_utils.go` to consume typed response from `elasticsearch.GetMLJobStats`
- [ ] 3.8 Update `internal/elasticsearch/ml/datafeed_state/*.go` to consume typed helpers for stats, start, and stop

## 4. Clean up redundant custom models

- [ ] 4.1 Remove `Datafeed`, `DatafeedCreateRequest`, `DatafeedUpdateRequest`, `DatafeedStats`, `DatafeedStatsResponse`, `MLJob`, and `MLJobStats` from `internal/models/ml.go`
- [ ] 4.2 Verify no other packages import the removed model structs
- [ ] 4.3 Run `go mod tidy` and `make build` to confirm everything compiles

## 5. Verify behavior and tests

- [ ] 5.1 Run `make check-lint` and fix any issues
- [ ] 5.2 Run `make build` to ensure no compilation errors
- [ ] 5.3 Run ML unit tests: `go test ./internal/elasticsearch/ml/...`
- [ ] 5.4 Run ML acceptance tests against a live Elasticsearch cluster
- [ ] 5.5 Verify error handling for 404/not-found scenarios remains unchanged
- [ ] 5.6 Verify force flags, timeouts, and optional parameters are still passed correctly
