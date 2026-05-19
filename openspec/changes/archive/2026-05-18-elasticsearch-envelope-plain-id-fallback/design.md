## Context

Canonical requirements for the envelope live in [`openspec/specs/entitycore-resource-envelope/spec.md`](../../specs/entitycore-resource-envelope/spec.md). Implementation lives in [`internal/entitycore/resource_envelope.go`](../../../internal/entitycore/resource_envelope.go).

The `KibanaResource` envelope in [`internal/entitycore/kibana_resource_envelope.go`](../../../internal/entitycore/kibana_resource_envelope.go) already handles this correctly via `resolveResourceIdentity` (line 173): it attempts `CompositeIDFromStr` and, if the result is `nil`, falls back to `GetResourceID()` / `GetSpaceID()`.

The `ElasticsearchResource` envelope has two places that hard-fail on non-composite IDs:

1. **`resolveElasticsearchReadResourceID`** (line 178): called by `Read` with `writeFallback == ""`. The final branch calls `clients.CompositeIDFromStr` and propagates its error diagnostic.
2. **`Delete`** (line 430): calls `clients.CompositeIDFromStr` directly and propagates its error diagnostic.

## Goals / Non-Goals

**Goals:**

- Add a lenient-resolution fallback to `resolveElasticsearchReadResourceID` that, when `CompositeIDFromStr` returns `nil`, falls back to `GetResourceID()` and, if that is empty, the raw ID string.
- Add the same fallback to `Delete` so it uses `GetResourceID()` / raw ID when the ID is not composite.
- Update unit tests in `resource_envelope_test.go` that assert on the strict-composite-error path to assert on the fallback happy path instead.
- Add an acceptance test `TestAccResourceAnomalyDetectionJobFrom0_12_2` that simulates a job originally stored with a plain `job_id` ID.

**Non-goals:**

- Changing the signature or behavior of `clients.CompositeIDFromStr`.
- Altering the composite-ID write path in `runWrite` (which uses `writeFallback` set to the non-empty `WriteID` string and is unaffected).
- Changing behavior for resources that already have composite IDs in state.
- Implementing import in resources that do not already declare `ImportState`.

## Decisions

- **Fallback order**: `CompositeIDFromStr` → `GetResourceID()` → raw `GetID()` string. This mirrors the KibanaResource pattern and ensures that both old-state scenarios (where `job_id` is populated) and plain-import scenarios (where `job_id` may be null) are covered.
- **No error diagnostic on plain ID**: A plain ID is treated as valid input, not as an error. The existing empty-string guard (`if resourceID == ""`) continues to fire when resolution produces nothing useful.
- **Delete path**: Replace the direct `clients.CompositeIDFromStr` call with the same three-step fallback so Delete handles plain IDs symmetrically with Read.
- **Unit tests**: The existing unit tests that assert `diags.HasError()` for a plain ID should become fallback happy-path assertions. Any tests covering the composite path should be retained.
- **Acceptance test**: The acceptance test for `TestAccResourceAnomalyDetectionJobFrom0_12_2` simulates a v0.12.2 state fixture (plain `job_id` as `id`) and verifies that the current provider can refresh, plan, and apply using that state.

## Risks / Trade-offs

- **Side-effect on resources that always expected composite IDs**: If a resource receives a non-composite ID that happens to match an unrelated resource's name, the fallback could silently issue API calls against the wrong resource. This risk is minimal because import IDs are under operator control.
- **Acceptance test environment**: The test requires a live Elasticsearch cluster. It is gated by `TF_ACC` per standard convention.

## Open Questions

- None.
