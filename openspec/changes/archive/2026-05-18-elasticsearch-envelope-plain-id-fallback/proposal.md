## Why

The `ElasticsearchResource` envelope's `Read` and `Delete` preludes require the persisted `id` to be in composite format (`<cluster_uuid>/<resource_id>`). When a resource was created or imported using an older provider version that stored plain IDs — or when a user runs `terraform import` with only the resource identifier — the envelope hard-fails with:

```
Wrong resource ID.
Resource ID must have following format: <cluster_uuid>/<resource identifier>
```

This makes any such resource unmanageable (no refresh, no destroy) without manually editing Terraform state.

The `KibanaResource` envelope already handles this correctly via `resolveResourceIdentity`, which silently discards the composite-parse error and falls back to `GetResourceID()`. The `ElasticsearchResource` envelope needs the same fallback.

## What Changes

- **`internal/entitycore/resource_envelope.go`**: Update `resolveElasticsearchReadResourceID` so that when the ID is not composite it falls back to `GetResourceID()` (and, if that is empty, to the raw ID string) instead of returning an error diagnostic. Update `Delete` similarly, replacing the hard-fail `CompositeIDFromStr` call with a lenient resolution that falls back to `GetResourceID()` / raw ID.
- **`internal/entitycore/resource_envelope_test.go`**: Replace tests that assert on the strict-composite error path with fallback happy-path tests covering plain-ID state for both `Read` and `Delete`.
- **Acceptance test** (`internal/elasticsearch/ml/anomalydetectionjob/`): Add `TestAccResourceAnomalyDetectionJobFrom0_12_2` to verify that a job whose state was imported with a plain `job_id` (as produced by provider ≤ 0.12.2) is successfully refreshed, planned, and applied by the current provider.
- **OpenSpec**: Delta under `specs/entitycore-resource-envelope/spec.md` replaces the strict-composite Delete scenario and the strict-composite Read fallback scenario with lenient-resolution wording.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`entitycore-resource-envelope`**: `Read` and `Delete` preludes gain a lenient-ID-resolution path that falls back from composite parsing to `GetResourceID()` / raw ID, matching the `KibanaResource` pattern.

## Impact

- **Users**: Resources imported or created with plain IDs by older provider versions become manageable again without state surgery.
- **Code**: `internal/entitycore/resource_envelope.go`, `internal/entitycore/resource_envelope_test.go`, and one new acceptance test file in the anomaly detection job package.
- **Compatibility**: No schema change, no state upgrade required. Composite-ID state continues to work as before. The change is additive: strict composite parsing still succeeds; the fallback fires only when parsing returns `nil`.
