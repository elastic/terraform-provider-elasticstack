## 1. Resource Implementation

- [x] 1.1 Add `resource.ResourceWithImportState` to the `var _` compile-time interface check in `internal/elasticsearch/enrich/resource.go`
- [x] 1.2 Implement `ImportState` method on `enrichPolicyResource`: call `resource.ImportStatePassthroughID` to copy the import ID into `id`, then set `execute = true` via `resp.State.SetAttribute`

## 2. Acceptance Tests

- [x] 2.1 Add an `ImportState: true, ImportStateVerify: true` step to the existing enrich policy acceptance test in `internal/elasticsearch/enrich/acc_test.go`, using `ImportStateIdFunc` to return `<cluster_uuid>/<policy_name>` (matching the pattern from `internal/elasticsearch/cluster/script/acc_test.go`)

## 3. Verification

- [x] 3.1 Run `make build` to confirm the provider compiles with no errors
- [x] 3.2 Run the enrich policy acceptance tests to confirm the import step passes
