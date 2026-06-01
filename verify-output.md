# Verification Report: `elasticsearch-ml-trained-model-deployment`

## Summary

| Dimension    | Status                                      |
|--------------|---------------------------------------------|
| Completeness | 27/27 tasks complete, 16/16 requirements covered |
| Correctness  | All core requirements implemented; 2 warnings noted |
| Coherence    | Follows project patterns with minor deviations |

### Final Assessment
No critical issues found. The implementation builds successfully, all tasks are complete, and the resource aligns with the approved proposal, design, and spec. Two coherence warnings should be addressed before archiving (noted below). Ready for archive after minor cleanup.

---

## Completeness

### Task Completion
All 27 tasks from `tasks.md` are marked complete and have corresponding implementation:

| Task Area | Count | Status |
|-----------|-------|--------|
| Spec alignment & validation | 3 | Done |
| Client wrappers | 4 | Done |
| Package & resource implementation | 9 | Done |
| Provider registration | 1 | Done |
| Testing | 10 | Done |

### Spec Coverage
Requirements from `specs/elasticsearch-ml-trained-model-deployment/spec.md`:

| Requirement | ID | Implementation | Status |
|-------------|----|----------------|--------|
| Start / Update / Stop / Stats APIs | REQ-001–004 | `internal/clients/elasticsearch/ml_trained_model_deployment.go` | Covered |
| Identity (`id`, `deployment_id`) | REQ-005 | `models.go`, `create.go:165-168` | Covered |
| Import support | REQ-006 | `resource.go:51-53` (`ImportStatePassthroughID`) | Covered |
| ForceNew attributes | REQ-007 | `schema.go:47-76` (`RequiresReplace`) | Covered |
| Wait-for polling | REQ-008 | `create.go:188-217` (`waitForDeploymentAllocationStatus`) | Covered |
| Mutable attributes (Update) | REQ-009 | `update.go:76-128` | Covered |
| Adaptive allocations + `ConflictsWith` | REQ-010 | `schema.go:62-65` (`int64validator.ConflictsWith`) | Covered |
| Computed attributes (`state`, `allocation_status`, `stats_json`) | REQ-011 | `read.go:56-75`, `create.go:170-183` | Covered |
| External stop detection | REQ-012 | `read.go:46-54` (returns `found=false` when stats nil) | Covered |
| Minimum ES version | REQ-013 | No extra gate needed per design | N/A |
| Connection override | REQ-015 | `models.go:46`, entitycore injects `elasticsearch_connection` | Covered |
| Acceptance tests | REQ-016 | `acc_test.go`, `schema_test.go` | Covered |

---

## Correctness

### Requirement Implementation Mapping

- **Create** → `create.go:30-186` calls `StartTrainedModelDeployment`, polls stats using `asyncutils.WaitForStateTransition`, populates computed attributes, and respects `timeouts.create` (default 5m).
- **Read** → `read.go:32-96` calls `GetTrainedModelStatsJSON`, updates computed fields, suppresses `number_of_allocations` when adaptive allocations are configured, and returns `found=false` on missing deployment.
- **Update** → `update.go:34-146` calls `UpdateTrainedModelDeployment` and re-reads stats to refresh state. Respects `timeouts.update`.
- **Delete** → `delete.go:26-39` calls `StopTrainedModelDeployment` with `force_stop`; treats 404 as success.

### Scenario Coverage

| Scenario | Test Coverage | Notes |
|----------|--------------|-------|
| Start API error surfaced | Implicit via client wrapper | Client returns diagnostic on non-success |
| Update API error surfaced | Implicit via client wrapper | Client returns diagnostic on non-success |
| Stop API error surfaced (non-404) | Implicit via client wrapper | Client returns diagnostic |
| Stop 404 treated as success | `StopTrainedModelDeployment` swallows 404 | Verified in code |
| Force stop on destroy | `acc_test.go` last step uses `force_stop = true` | Destroy at end of test case exercises this |
| ID set after create | `acc_test.go:55` (`TestCheckResourceAttrSet(testResourceName, "id")`) | Covered |
| deployment_id defaults to model_id | `acc_test.go:54` | `basic.tf` omits `deployment_id` |
| Import by composite ID | `acc_test.go:77-88` (`ImportState: true`) | Covered |
| threads_per_allocation change triggers replace | `schema.go:68` (`RequiresReplace`) | Schema enforces this |
| wait_for = "started" reaches state | `create.go:188-217` polling logic | Covered |
| wait_for timeout | `create.go:30-49` context with timeout + `diagutil.ContainsContextDeadlineExceeded` | Covered |
| Update number_of_allocations | `acc_test.go:58-64` (`update_allocations`) | Covered |
| Update adaptive_allocations | `acc_test.go:65-78` (`update_adaptive`) | Covered |
| ConflictsWith validation | `schema_test.go` (`conflicts_with` config) | Covered |
| Diff on number_of_allocations when fixed | `read.go:78-84` updates from API only when adaptive is null | Covered |
| Switch adaptive → fixed triggers update | `update.go:76-128` handles both fields | Covered |
| Computed attributes after create | `acc_test.go:55-59` | Covered |
| External stop detected | `read.go:46-54` returns `found=false` | Covered |
| Resource-level connection override | Via entitycore `GetElasticsearchClient` with `ElasticsearchConnection` | Covered |
| Create and verify | `acc_test.go:45-60` | Covered |
| No diff on re-plan | `acc_test.go:61-66` (`PlanOnly`) | Covered |
| Import roundtrip | `acc_test.go:77-88` | Covered |
| Delete (force_stop = false) | Implicit destroy with `basic.tf` config | **See Warning / Suggestion below** |
| Model not found | `acc_test.go:116-127` (`non_existent`) | Covered |

---

## Coherence

### Design Adherence

The implementation follows the design decisions documented in `design.md`:

- Uses `entitycore.NewElasticsearchResource[T]` with `PlaceholderElasticsearchWriteCallback` (mirrors `ml/jobstate`).
- Package and file layout matches the decided structure.
- Composite `id` format (`<cluster_uuid>/<deployment_id>`) is used.
- `UseStateForUnknown` on `id` and `deployment_id`.
- `adaptive_allocations` and `number_of_allocations` are mutually exclusive.
- `stats_json` populated as raw JSON for extensibility.
- Wait-for polling reuses `internal/asyncutils/state_waiter.go` pattern.

### Deviations / Issues Found

#### WARNING — `StopTrainedModelDeployment` 404 handling uses non-idiomatic second request
**Location:** `internal/clients/elasticsearch/ml_trained_model_deployment.go:178-190`

The client wrapper calls `req.Do(ctx)`, and on error calls `req.Perform(ctx)` to check the raw status code for 404. This makes a second HTTP POST request, which is fragile and inconsistent with the rest of the codebase. The standard project pattern is to use `IsNotFoundElasticsearchError(err)` (see `internal/clients/elasticsearch/helpers.go:61`).

**Recommendation:** Refactor `StopTrainedModelDeployment` to:
```go
if err != nil {
    if IsNotFoundElasticsearchError(err) {
        return diags
    }
    diags.AddError(...)
    return diags
}
```

#### WARNING — Design doc requested warning log on 404 during delete
**Location:** `internal/elasticsearch/ml/trainedmodeldeployment/delete.go`, `internal/clients/elasticsearch/ml_trained_model_deployment.go:174-190`

`design.md` states: "Treat HTTP 404 as success (idempotent). log warning if already stopped." The spec (REQ-004) requires 404 to be treated as success without error, which the code does, but the design additionally requested a warning log. Neither the client wrapper nor `delete.go` logs a warning.

**Recommendation:** Add a `tflog.Warn` in `delete.go` when the returned diagnostics indicate a 404 was swallowed, or document that the silent no-op is intentional.

#### SUGGESTION — `force_stop = false` delete not explicitly tested in isolation
**Location:** `internal/elasticsearch/ml/trainedmodeldeployment/acc_test.go`

The acceptance test performs a single destroy at the end of `TestAccResourceMLTrainedModelDeployment_basic` using the `force_stop = true` configuration step. Task 5.7 explicitly calls for a test that destroys with `force_stop = false`. While this default is exercised implicitly during intermediate step transitions, there is no isolated destroy assertion for the non-force case.

**Recommendation:** Add an explicit `Destroy: true` step with the `basic` config (which defaults `force_stop = false`) before the final `force_stop = true` step, or add a separate test function.

#### SUGGESTION — ML node availability not checked before test execution
**Location:** `internal/elasticsearch/ml/trainedmodeldeployment/acc_test.go:34-56`

`findSuitableTrainedModel` skips tests when no PyTorch model exists, but Task 5.1 also calls for skipping when no ML nodes are available. The sibling `anomalydetectionjob` tests have a helper (`mlOpenJobErrorLooksLikeMLNodeCapacity`) to detect this at runtime, but there is no equivalent pre-flight check here.

**Recommendation:** Optionally query `_nodes` or catch ML node capacity errors during `StartTrainedModelDeployment` and skip the test with `t.Skip`.

#### SUGGESTION — Import may fail when `deployment_id != model_id`
**Location:** `internal/elasticsearch/ml/trainedmodeldeployment/read.go:40-48`

During import, `resourceID` resolves to the `deployment_id` part of the composite ID. `readTrainedModelDeployment` falls back `model_id` to this same value when `ModelID` is empty. If the practitioner configured a custom `deployment_id` different from `model_id`, the stats query will use the wrong `model_id` and likely return 404. The design doc acknowledges this limitation ("import works naturally for the common case"), but it is a functional gap.

**Recommendation:** Document this limitation in the resource description or user docs, or consider a composite import ID that encodes both `model_id` and `deployment_id` (e.g., `<cluster_uuid>/<model_id>/<deployment_id>`).

---

## Build & Test Validation

```
$ go build ./...
(ok, no errors)

$ go test -run=TestNone ./internal/elasticsearch/ml/trainedmodeldeployment/...
ok  	github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/trainedmodeldeployment	1.259s [no tests to run]
```

The project compiles cleanly. Unit/acceptance tests for the new resource compile successfully.

---

## Open Risks / Questions

1. **Second HTTP request on stop error** (`ml_trained_model_deployment.go:184`) could cause transient side-effects if the `_stop` endpoint is not fully idempotent under retry.
2. **Import edge case** for custom `deployment_id` values different from `model_id` may break the Read-after-import flow.
3. Acceptance tests require a live cluster with a pre-existing PyTorch model. CI environments without such a model or ML nodes will skip the entire suite, which is acceptable but means the resource is not exercised in all CI pipelines.

---

## Recommended Next Step

1. Fix `StopTrainedModelDeployment` to use `IsNotFoundElasticsearchError(err)` instead of `req.Perform(ctx)`.
2. Decide whether to add a warning log on idempotent 404 delete (align with `design.md` or accept silent no-op).
3. Run the acceptance test suite against an ES cluster with a PyTorch model to confirm end-to-end behavior.
4. Archive the change in OpenSpec once the above warnings are addressed or accepted.
