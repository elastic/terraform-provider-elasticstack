## Why

The Elasticsearch `searchable_snapshot` ILM action supports `force_merge_on_clone` (bool, default `true`, GA since 9.2) alongside `force_merge_index` and `snapshot_repository`. It controls whether the `force_merge_index` merge runs on a temporary clone of the managed index (the default) or directly on the index itself. Per elastic/elasticsearch#133954, if the clone-and-force-merge sequence is interrupted (busy master, cluster load) before the clone is promoted and cleaned up, the resulting snapshot can carry a permanent `index.lifecycle.skip: true` setting. When later restored, the mounted index gets stuck outside any ILM phase and requires manually clearing `index.lifecycle.skip` via `_settings` before ILM evaluates it again. Setting `force_merge_on_clone: false` skips the clone step and avoids this failure mode for policies that don't need the force-merge (e.g. already-optimized restored data).

`elasticstack_elasticsearch_index_lifecycle`'s `searchable_snapshot` block (used identically in `hot`, `cold`, and `frozen`) only exposes `snapshot_repository` and `force_merge_index` today, so `force_merge_on_clone` cannot be set from Terraform and is stuck at the Elasticsearch default.

## What Changes

- Add `force_merge_on_clone` (bool, optional + computed, default `true`) to the `searchable_snapshot` action block, available in `hot`, `cold`, and `frozen` phases (same block shape everywhere).
- Gate `force_merge_on_clone` behind a `9.2.0` minimum-version check using the existing `ilmActionSettingOptions` version-gating mechanism (same pattern as `shrink.allow_write_after_shrink`'s `8.14.0` gate): a non-default value on an unsupported server returns an error diagnostic; the default value is silently omitted from the payload.
- Change the ILM policy write path (`PutIlm` in `internal/clients/elasticsearch/ilm.go`) to submit the marshaled policy JSON via the typed client's `.Raw()` body instead of unmarshaling into `putlifecycle.Request` and using `.Request()`. This is required because the vendored `go-elasticsearch` typed struct for the searchable-snapshot action (and the upstream `elasticsearch-specification` it is generated from) does not model `force_merge_on_clone` yet; round-tripping through the typed struct would silently drop the field before it reaches Elasticsearch. `ClearILMPolicyFromIndices` in the same file already uses this `.Raw()` pattern for the same typed-struct-gap reason, so this is not a new pattern in this package.
- Extend `TestAccResourceILMSearchableSnapshotPhases` with an additional `9.2`-gated test step (not a new test function) exercising `force_merge_on_clone`, following the existing `SkipFunc: versionutils.CheckIfVersionIsUnsupported(...)` convention used elsewhere in `acc_test.go`.
- File an upstream tracking issue against `elastic/elasticsearch-specification` so the `.Raw()` bypass in `PutIlm` can be reverted once the typed struct models `force_merge_on_clone` (follow-up task, not a blocker for this change).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-index-lifecycle`: add a new version-gated `searchable_snapshot.force_merge_on_clone` setting (extends REQ-022–REQ-025's version-gating pattern) and document the `PutIlm` write-path change from the typed `.Request()` call to `.Raw()`.

## Impact

- `internal/elasticsearch/index/ilm/constants.go` — add `attrForceMergeOnClone`.
- `internal/elasticsearch/index/ilm/schema_actions.go` — add the attribute to `blockSearchableSnapshot()` and `blockSearchableSnapshotInFrozenPhase()`.
- `internal/elasticsearch/index/ilm/attr_types.go` — add the attribute to `searchableSnapshotObjectType()`.
- `internal/elasticsearch/index/ilm/expand.go` — register the setting in `ilmActionSettingOptions` (`def: true`, `minVersion: 9.2.0`) and wire it into the `ilmActionSearchableSnapshot` case of `expandPhase`.
- `internal/clients/elasticsearch/ilm.go` — swap `PutIlm` to use `.Raw()` instead of `.Request(&putlifecycle.Request{...})`; drop the now-unused typed unmarshal (and the `putlifecycle` import) if nothing else in the function needs it.
- `internal/elasticsearch/index/ilm/acc_test.go` — add a package-level `searchableSnapshotForceMergeOnCloneMinSupportedVersion = version.Must(version.NewVersion("9.2.0"))` near the other version vars, plus a new gated step in `TestAccResourceILMSearchableSnapshotPhases` and any needed `testdata` fixtures.
- `openspec/specs/elasticsearch-index-lifecycle/spec.md` — extend REQ-022–REQ-025 (or add a new requirement) and the `searchable_snapshot` schema block documentation; update generated provider docs for the resource.
- Follow-up (outside this PR): file an issue against `elastic/elasticsearch-specification` requesting `force_merge_on_clone` support in the generated `SearchableSnapshotAction` type.
