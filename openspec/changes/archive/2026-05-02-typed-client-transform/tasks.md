## 1. Typed client migration — transform helpers

- [x] 1.1 Rewrite `PutTransform` in `internal/clients/elasticsearch/transform.go` to use `typedClient.Transform.PutTransform(name).Raw(body).Timeout(...).DeferValidation(...).Do(ctx)`
- [x] 1.2 Rewrite `GetTransform` to use `typedClient.Transform.GetTransform().TransformId(name).Perform(ctx)` and manual decode into `models.GetTransformResponse`
- [x] 1.3 Rewrite `GetTransformStats` to use `typedClient.Transform.GetTransformStats(name).Do(ctx)` and search `[]types.TransformStats` for the matching ID
- [x] 1.4 Rewrite `UpdateTransform` to use `typedClient.Transform.UpdateTransform(name).Raw(body).Timeout(...).DeferValidation(...).Do(ctx)`
- [x] 1.5 Rewrite `DeleteTransform` to use `typedClient.Transform.DeleteTransform(name).Force(true).Do(ctx)`
- [x] 1.6 Rewrite `startTransform` to use `typedClient.Transform.StartTransform(name).Timeout(...).Do(ctx)`
- [x] 1.7 Rewrite `stopTransform` to use `typedClient.Transform.StopTransform(name).Timeout(...).Do(ctx)`

## 2. Resource and test updates

- [x] 2.1 Update `internal/elasticsearch/transform/transform.go` to call the migrated helpers (verify signatures remain compatible)
- [x] 2.2 Update `internal/elasticsearch/transform/transform_test.go` for any signature or type changes
- [x] 2.3 Verify all transform testdata configurations still compile and run correctly

## 3. Model cleanup

- [x] 3.1 Remove `models.PutTransformParams` from `internal/models/transform.go` once `PutTransform` no longer needs it
- [x] 3.2 Remove `models.UpdateTransformParams` from `internal/models/transform.go` once `UpdateTransform` no longer needs it
- [x] 3.3 Remove `models.TransformStats` and `models.GetTransformStatsResponse` from `internal/models/transform.go` once `GetTransformStats` returns `*types.TransformStats` directly
- [x] 3.4 Verify `models.Transform` and `models.GetTransformResponse` are not removed — they are still needed for `.Raw()` body construction and manual response decode

## 4. Build and testing

- [x] 4.1 Run `make build` to confirm compilation
- [x] 4.2 Run unit tests for `internal/elasticsearch/transform`
- [x] 4.3 Run acceptance tests for `elasticstack_elasticsearch_transform` (run on CI; skipped locally due to no running Elastic Stack)
- [x] 4.4 Run `make check-lint` and `make check-openspec`
