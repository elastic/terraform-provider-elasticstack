## 1. Shared guard

- [x] 1.1 Decide the exact shared-helper shape (or confirm per-package inline guards), mirroring `internal/elasticsearch/index/template/write.go:56-65`: compute `id` via `client.ID` only when there is no prior state; otherwise carry `id` forward unchanged.
- [x] 1.2 If a shared helper is added, place it where all four packages can use it without an import cycle (likely `internal/entitycore` or a small shared utility package) and unit test the helper in isolation.

## 2. security_role

- [x] 2.1 In `writeRole` (`internal/elasticsearch/security/role/update.go`), only call `client.ID` when `req.Prior == nil`; when `req.Prior != nil`, set `data.ID = req.Prior.ID` and skip the `client.ID` call entirely.
- [x] 2.2 Add a unit test where `req.Prior.ID` carries a cluster-UUID prefix that does not match what a (mocked) `client.ID` call would currently produce, asserting the returned `id` equals `req.Prior.ID` unchanged and that Update still succeeds.
- [x] 2.3 Add an acceptance test that imports a role with a deliberately incorrect cluster-UUID prefix in its `id`, then applies a config change to a role attribute, and asserts the apply succeeds without an inconsistent-result error and without `id` changing.

## 3. data_stream_lifecycle

- [x] 3.1 In `writeDataStreamLifecycle` (`internal/elasticsearch/index/datastreamlifecycle/write.go`), apply the same `req.Prior == nil` guard as `security_role`.
- [x] 3.2 Add a unit test mirroring 2.2 for this resource.

## 4. watch

- [x] 4.1 In `updateWatch` (`internal/elasticsearch/watcher/watch/update.go`), remove the `client.ID` call and the recompute of `plan.ID`; since this function only runs for Update (`createWatch` is separate), preserve the existing `id` from prior state unconditionally.
- [x] 4.2 Confirm `createWatch` (`internal/elasticsearch/watcher/watch/create.go`) is unaffected and still computes `id` via `client.ID` on Create.
- [x] 4.3 Add a unit test mirroring 2.2 for this resource.

## 5. ml_datafeed

- [ ] 5.1 In `updateDatafeed` (`internal/elasticsearch/ml/datafeed/update.go`), remove the `client.ID` call and the recompute of `plan.ID`; preserve the existing `id` from prior state unconditionally.
- [ ] 5.2 Confirm `createDatafeed` (`internal/elasticsearch/ml/datafeed/create.go`) is unaffected and still computes `id` via `client.ID` on Create.
- [ ] 5.3 Add a unit test mirroring 2.2 for this resource.

## 6. Spec sync

- [ ] 6.1 Update `openspec/specs/elasticsearch-security-role/spec.md` Requirement: Identity (REQ-005–REQ-006) per this change's delta spec.
- [ ] 6.2 Update `openspec/specs/elasticsearch-data-stream-lifecycle/spec.md` Requirement: Identity (REQ-005–REQ-006) per this change's delta spec.
- [ ] 6.3 Update `openspec/specs/elasticsearch-watch/spec.md` Requirement: Identity (REQ-005–REQ-006) per this change's delta spec.
- [ ] 6.4 Update `openspec/specs/elasticsearch-ml-datafeed/spec.md` Requirement: Identity and import (REQ-007–REQ-009) per this change's delta spec.

## 7. Verification

- [ ] 7.1 `make build`
- [ ] 7.2 `go test ./internal/elasticsearch/security/role/... ./internal/elasticsearch/index/datastreamlifecycle/... ./internal/elasticsearch/watcher/watch/... ./internal/elasticsearch/ml/datafeed/...`
- [ ] 7.3 Run the new/extended acceptance test(s) against a live stack (`TF_ACC=1`) per `dev-docs/high-level/testing.md`
- [ ] 7.4 `make check-openspec`
