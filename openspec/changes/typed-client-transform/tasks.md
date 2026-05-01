## 1. Typed client migration — transform helpers

- [ ] 1.1 Rewrite `PutTransform` in `internal/clients/elasticsearch/transform.go` to use `typedClient.Transform.PutTransform(name).Raw(body).Timeout(...).DeferValidation(...).Do(ctx)`
- [ ] 1.2 Rewrite `GetTransform` to use `typedClient.Transform.GetTransform().TransformId(name).Perform(ctx)` and manual decode into `models.GetTransformResponse`
- [ ] 1.3 Rewrite `GetTransformStats` to use `typedClient.Transform.GetTransformStats(name).Do(ctx)` and search `[]types.TransformStats` for the matching ID
- [ ] 1.4 Rewrite `UpdateTransform` to use `typedClient.Transform.UpdateTransform(name).Raw(body).Timeout(...).DeferValidation(...).Do(ctx)`
- [ ] 1.5 Rewrite `DeleteTransform` to use `typedClient.Transform.DeleteTransform(name).Force(true).Do(ctx)`
- [ ] 1.6 Rewrite `startTransform` to use `typedClient.Transform.StartTransform(name).Timeout(...).Do(ctx)`
- [ ] 1.7 Rewrite `stopTransform` to use `typedClient.Transform.StopTransform(name).Timeout(...).Do(ctx)`

## 2. Resource and test updates

- [ ] 2.1 Update `internal/elasticsearch/transform/transform.go` to call the migrated helpers (verify signatures remain compatible)
- [ ] 2.2 Update `internal/elasticsearch/transform/transform_test.go` for any signature or type changes
- [ ] 2.3 Verify all transform testdata configurations still compile and run correctly

## 3. Model cleanup

- [ ] 3.1 Remove `models.PutTransformParams` from `internal/models/transform.go` once `PutTransform` no longer needs it
- [ ] 3.2 Remove `models.UpdateTransformParams` from `internal/models/transform.go` once `UpdateTransform` no longer needs it
- [ ] 3.3 Remove `models.TransformStats` and `models.GetTransformStatsResponse` from `internal/models/transform.go` once `GetTransformStats` returns `*types.TransformStats` directly
- [ ] 3.4 Verify `models.Transform` and `models.GetTransformResponse` are not removed — they are still needed for `.Raw()` body construction and manual response decode

## 4. Build and testing

- [ ] 4.1 Run `make build` to confirm compilation
- [ ] 4.2 Run unit tests for `internal/elasticsearch/transform`
- [ ] 4.3 Run acceptance tests for `elasticstack_elasticsearch_transform`
- [ ] 4.4 Run `make check-lint` and `make check-openspec`
