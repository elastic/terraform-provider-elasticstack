## Context

Elasticsearch's `searchable_snapshot` ILM action accepts `snapshot_repository` (required), `force_merge_index` (default `true`), and `force_merge_on_clone` (default `true`, GA since Elasticsearch 9.2). `force_merge_on_clone` controls whether the `force_merge_index` merge happens on a temporary clone of the managed index (default behavior) or directly on the index. Per elastic/elasticsearch#133954, an interrupted clone-and-merge sequence can leave a snapshot permanently marked `index.lifecycle.skip: true`; setting `force_merge_on_clone: false` avoids that code path.

`elasticstack_elasticsearch_index_lifecycle`'s `searchable_snapshot` block (same shape in `hot`, `cold`, and `frozen`) currently only exposes `snapshot_repository` and `force_merge_index`. The action round-trips through a generic `map[string]any` (`models.Action`, `internal/models/index.go:100`) on the Terraform side, but the vendored `go-elasticsearch` client's typed struct for this action (`typedapi/types/searchablesnapshotaction.go`) does not model `force_merge_on_clone` either, and neither does the upstream `elasticsearch-specification` (`specification/ilm/_types/Phase.ts`, checked at HEAD of `main`). Terraform-only schema wiring would silently no-op against the real API: `PutIlm` (`internal/clients/elasticsearch/ilm.go:35-46`) marshals the generic policy to JSON, then unmarshals it into a typed `putlifecycle.Request` before calling `.Request(&req)`; the typed struct's `UnmarshalJSON` silently drops unrecognized keys, so `force_merge_on_clone` would be accepted into Terraform state but never reach Elasticsearch.

This design was produced by a prior implementation-research pass on the linked issue, which evaluated two approaches (below) and got explicit sign-off from `@tobio` on the recommended one, including the ES version floor and the acceptable scope of the `PutIlm` change.

## Goals / Non-Goals

**Goals:**

- Let practitioners set `force_merge_on_clone = false` on `hot`, `cold`, and `frozen` `searchable_snapshot` blocks, and have that value actually reach Elasticsearch's ILM policy.
- Gate the new setting on Elasticsearch `9.2.0`, consistent with the existing `ilmActionSettingOptions` version-gating mechanism (e.g. `attrAllowWriteAfterShrink` at `8.14.0`).
- Keep the default (`true`) behavior unchanged and never send the field when unsupported and left at its default.

**Non-Goals:**

- `replicate_for` / `total_shards_per_node` on `searchable_snapshot` — other documented options, not requested here.
- Changing the default of `force_merge_index` or `force_merge_on_clone` (both stay Elasticsearch's default of `true`).
- Any remediation tooling for indices already stuck with `index.lifecycle.skip: true`.
- A broader `PutIlm` / `GetIlm` refactor beyond the minimal `.Raw()` swap described below.
- Actually filing/landing the upstream `elasticsearch-specification` issue — tracked as a follow-up task, not part of this change's implementation.

## Approaches considered

These were evaluated during research prior to this proposal and are treated as settled; they are recorded here for context, not re-opened for exploration.

### Approach A — Terraform-only schema/expand wiring (rejected)

Mirror `force_merge_index`: add `attrForceMergeOnClone` as a `BoolAttribute` (Optional+Computed+Default `true`) in `blockSearchableSnapshot()` and `blockSearchableSnapshotInFrozenPhase()`, add it to `searchableSnapshotObjectType()`, include it in `expandAction(...)`.

**Rejected because** `PutIlm` unmarshals the generic policy into a typed `putlifecycle.Request` before calling `.Request(&req)`; the generated `types.SearchableSnapshotAction` only declares `ForceMergeIndex` / `SnapshotRepository`, and its `UnmarshalJSON` silently drops unrecognized keys. Upstream `elasticsearch-specification`'s `Phase.ts` lacks the field too, so this gap can't be closed by only touching Terraform code — Approach A would compile, store the value in state, and never reach Elasticsearch, i.e. a silent no-op.

### Approach B — Same Terraform wiring, plus bypass the typed struct for `PutIlm` (chosen)

Everything from Approach A, plus send the marshaled policy JSON via the raw body path instead of `.Request()`:

```go
_, err = typedClient.Ilm.PutLifecycle(policy.Name).Raw(bytes.NewReader(policyBytes)).Do(ctx)
```

`PutLifecycle.Raw(io.Reader)` exists on the typed client (`typedapi/ilm/putlifecycle/put_lifecycle.go:115`), and `.Do()` gives the raw body precedence over `.Request()` (same file, line 146) when both are set — in practice this means dropping the `.Request(&req)` call and the preceding typed unmarshal entirely, and only ever calling `.Raw(...)`. This is not a new pattern in this package: `ClearILMPolicyFromIndices` (`internal/clients/elasticsearch/ilm.go:118-135`) already uses `.Raw()` on `Indices.PutSettings()` for the identical reason (a typed-struct gap).

## Decisions

- **Bypass the typed request path in `PutIlm`.** Confirmed acceptable by `@tobio`: "Moving off the typed path is fine." This affects all ILM policy writes through this provider, not just `searchable_snapshot`, but the payload is still built from the same `models.Policy` marshaling used today, so this is a transport-layer change, not a data-shape change.
- **Version gate at `9.2.0`.** Confirmed by `@tobio`: the provider's supported Elasticsearch floor is `8.0+`, and the `9.2` minVersion gate on `force_merge_on_clone` is required (the field is GA only from `9.2` onward). Implemented via the existing `ilmActionSettingOptions` map with `def: true, minVersion: version.Must(version.NewVersion("9.2.0"))`, following the `attrAllowWriteAfterShrink` (`8.14.0`) precedent — a non-default (`false`) value on an unsupported server is a hard error diagnostic, not a silent drop.
- **Test as an extension of the existing test, not a new one.** Confirmed by `@tobio`: add a new `SkipFunc`-gated step to `TestAccResourceILMSearchableSnapshotPhases` rather than a separate test function, following the `searchableSnapshotForceMergeUpdateVersionLimit` / line-425 convention already in `acc_test.go`.
- **File the upstream issue as a follow-up, not a blocking task.** The `.Raw()` bypass is intentionally temporary; once `elasticsearch-specification` and the generated `go-elasticsearch` client model `force_merge_on_clone`, `PutIlm` can revert to the typed `.Request()` path. Filing that issue is tracked in `tasks.md` but does not block merging this change's implementation.

## Risks / Trade-offs

- **[Risk] `PutIlm` losing typed-request validation for all ILM writes**, not just `searchable_snapshot`. Mitigation: the payload is still constructed from the same `models.Policy` → JSON marshal used today; only the transport call changes from `.Request()` to `.Raw()`. Elasticsearch itself still validates the payload server-side, same as any other `.Raw()` call in this codebase (e.g. `ClearILMPolicyFromIndices`).
- **[Risk] This is a workaround for an upstream gap that could resurface if `elasticsearch-specification` changes shape** before the field is added. Mitigation: filing the upstream tracking issue (see `tasks.md`) creates a paper trail to revert the `.Raw()` bypass once upstream support lands.
- **[Trade-off] Version-gating adds one more entry to `ilmActionSettingOptions`** that must be kept in sync with Elasticsearch's own version support matrix. This mirrors the existing pattern for `rollover.max_primary_shard_docs` (`8.2.0`), the `rollover.min_*` settings (`8.4.0`), and `shrink.allow_write_after_shrink` (`8.14.0`), so it is consistent with how the provider already manages this class of risk.

## Open questions

None blocking — all four questions from the prior research run were answered by `@tobio`. Two small implementation-detail items remain:

- File the upstream issue as part of this PR's checklist, or track separately?
- Does dropping the typed `putlifecycle.Request` unmarshal in `PutIlm` need any error-handling adjustment for malformed policies, or is it safe to remove outright?
