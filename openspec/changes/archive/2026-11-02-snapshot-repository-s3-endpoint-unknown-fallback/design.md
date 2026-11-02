## Context

`elasticstack_elasticsearch_snapshot_repository` reads an S3 repository back through `settingsToS3` (`internal/elasticsearch/snapshot/repository/read.go`). Elasticsearch does not return `endpoint` in `GET /_snapshot/{name}` when it was never set, so REQ-017 makes the read inherit the prior `endpoint` value when the API omits it. The provider uses the same read for refresh, import, and the read that follows create and update. In the read after create or update, the prior value is the planned value.

`s3.endpoint` is Optional + Computed with no plan modifier. When it is not configured, the framework plans it as unknown on create and on every update. The current inheritance only checks whether the prior value is null, so it carries the unknown value into state. Terraform then rejects the apply result.

The neighbouring fallbacks already ignore unknown prior values. `path_style_access` checks `typeutils.IsKnown` before inheriting. The common settings (`chunk_size`, `max_snapshot_bytes_per_sec`, `max_restore_bytes_per_sec`) go through `strSettingNullWithFallback`, which does the same.

## Goals / Non-Goals

**Goals:**

- State never holds an unknown `s3.endpoint` after create, update, refresh, or import.
- Create and update of an S3 repository without `endpoint` succeed, and a re-plan is empty.
- A configured `endpoint` that Elasticsearch does not echo is still preserved (REQ-017 unchanged for known values).

**Non-Goals:**

- No schema change to `s3.endpoint` (no plan modifier, no change to Optional + Computed).
- No change to `path_style_access` or the common-setting fallbacks.
- No change to removing a previously configured `endpoint` on its own: Optional + Computed keeps the prior value in state. When the removal is applied together with another change, the planned value is unknown and the read-back guard now yields null instead of a failed apply.
- No change to the data source.

## Decisions

**Guard the inheritance with a known-value check in the read-back.** The `endpoint` fallback in `settingsToS3` inherits the prior value only when `typeutils.IsKnown` reports it as known. An unknown or null prior value with no `endpoint` in the GET response maps to null. This puts `endpoint` on the same known-value rule as `path_style_access` and `strSettingNullWithFallback`, so all S3 read-back fallbacks ignore unknown prior values. The `endpoint` branch keeps its existing treatment of an empty API value as omitted, so it is not replaced by `strSettingNullWithFallback`. The API value still wins whenever Elasticsearch returns a non-empty `endpoint`. `path_style_access` has a schema default and is never unknown in a plan, so its rule is stated for completeness and covered as a characterization case.

`stringplanmodifier.UseStateForUnknown()` on `endpoint` was considered instead. It only carries a prior state value into the plan, so it would fix update but not create, where there is no prior state and the planned value stays unknown. It would also change plan output for an attribute whose read-back is the actual defect. The read-back guard fixes both create and update without touching the schema.

**Scope of the change is the read mapping only.** The write path already omits an empty `endpoint` from the PUT body (`setIfNotEmpty`), and the schema already validates configured values. Nothing outside `settingsToS3` changes.

**Testing approach.**

- Unit: extend the existing table test for S3 state inheritance with an unknown prior `endpoint` and a GET response without `endpoint`. This covers the exact read-after-write input without a cluster.
- Acceptance: a dedicated test that creates an S3 repository without `endpoint`, re-plans the unchanged configuration, updates another S3 setting, sets `endpoint`, and finally removes `endpoint` together with another change. It runs against the standard acceptance stack. `verify = false` means no S3 bucket or credentials are needed, like the existing S3 acceptance test. The existing S3 acceptance test, which always sets `endpoint`, keeps covering the known-value path.

## Risks / Trade-offs

- [A configured `endpoint` could be read back as unknown in some other path and now become null instead of failing] → Configured values are always known in the plan, and refresh and import use known prior state, so only the unconfigured read after create or update takes the new branch. The known-value scenarios in REQ-017 and the existing S3 acceptance test guard the old behavior.
- [Users who set `endpoint` as a workaround and then remove it see no plan change] → Intended Optional + Computed behavior: the value stays in state and in Elasticsearch until an apply that also changes another attribute, which then clears it. Documented in the proposal; making a lone removal clear the value is out of scope.

## Migration Plan

None. Existing state is unaffected, and repositories that failed before apply cleanly after upgrading. A configuration that set `endpoint` only to work around this bug can keep it; removing it takes effect on the next apply that also changes another attribute.

## Open Questions

None.
