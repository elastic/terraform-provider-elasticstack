## 1. Envelope implementation

- [x] 1.1 Update `resolveElasticsearchReadResourceID` in `internal/entitycore/resource_envelope.go`: when `CompositeIDFromStr` returns `nil` (non-composite ID), fall back to `model.GetResourceID().ValueString()`; if that is empty, fall back to the raw `model.GetID().ValueString()`. Remove the error-diagnostic propagation from the `CompositeIDFromStr` call in this function.
- [x] 1.2 Update `ElasticsearchResource.Delete` in `internal/entitycore/resource_envelope.go`: replace the direct `clients.CompositeIDFromStr` call (which hard-fails on plain IDs) with the same three-step fallback as 1.1 — composite parse → `GetResourceID()` → raw ID.

## 2. Unit tests

- [x] 2.1 Update `internal/entitycore/resource_envelope_test.go`: change any test that currently asserts a hard-fail diagnostic for a plain (non-composite) `id` in `Read` to instead assert that `Read` succeeds and invokes the `readFunc` with the plain resource ID.
- [x] 2.2 Update `internal/entitycore/resource_envelope_test.go`: change any test that currently asserts a hard-fail diagnostic for a plain (non-composite) `id` in `Delete` to instead assert that `Delete` succeeds and invokes the `deleteFunc` with the plain resource ID.
- [x] 2.3 Ensure existing composite-ID unit tests still pass (no regression on the composite path).

## 3. Acceptance test

- [x] 3.1 Add `TestAccResourceAnomalyDetectionJobFrom0_12_2` in `internal/elasticsearch/ml/anomalydetectionjob/acc_test.go`. The test SHALL:
  - Create an ML anomaly detection job using provider `v0.12.2`.
  - Import the resource with a plain `job_id` (not `cluster-id/job-id`) using the old provider, persisting the state.
  - Switch to the current provider and apply the same config — confirming refresh, plan, and apply all succeed without "Wrong resource ID" diagnostics.
  - Assert that the provider successfully refreshes (`terraform plan`) and applies (`terraform apply`) without errors.
- [x] 3.2 Add any required testdata directory and configuration files under `internal/elasticsearch/ml/anomalydetectionjob/testdata/TestAccResourceAnomalyDetectionJobFrom0_12_2/` if the test framework requires static `.tf` files.

## 4. OpenSpec

- [x] 4.1 Keep delta spec `openspec/changes/elasticsearch-envelope-plain-id-fallback/specs/entitycore-resource-envelope/spec.md` aligned with the implementation once tasks 1–3 are complete.
- [ ] 4.2 After merge decision: **sync** into `openspec/specs/entitycore-resource-envelope/spec.md` or **archive** the change per project workflow; run `make check-openspec`.
