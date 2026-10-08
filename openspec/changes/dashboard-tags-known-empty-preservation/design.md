## Context

`elasticstack_kibana_dashboard` declares `tags` as an `Optional` `schema.ListAttribute` of strings
(`internal/kibana/dashboard/schema.go:157-161`). The read-back mapping in
`dashboardPopulateFromAPI` (`internal/kibana/dashboard/models.go:104-109`) is:

```go
// models.go:104-109
if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
    m.Tags = typeutils.SliceToListTypeString(ctx, *data.Data.Tags, path.Root("tags"), &diags)
} else {
    m.Tags = types.ListNull(types.StringType)
}
```

Any nil-or-empty API `tags` value collapses to `types.ListNull`, which mismatches a known-empty
`[]` plan value and produces the reported
`.tags: was cty.ListValEmpty(cty.String), but now null` inconsistent-result error. The
create/update request mapping (`models.go:178-183`, `222-227`) already sends an empty array
correctly; the bug is isolated to this read-back branch.

This is the same symptom class already fixed for the dashboard root `description` field
(`models.go:49-65`, REQ-009, archived change `2026-07-02-dashboard-description-null-normalization`)
and for `time_range.mode` (`models.go:67-77`). It is also the same class already fixed for `tags`
specifically on two sibling resources that model `tags` as `types.Set`:
`internal/kibana/security_exception_item/models.go:595-597` and
`internal/kibana/agentbuildertool/models.go:112`, both via
`typeutils.SetFromAPIStringsPreserveKnownEmpty` (`internal/utils/typeutils/tfsdk_sets.go:60-68`).
Dashboard models `tags` as `types.List`, not `types.Set`, so that exact helper does not apply
as-is.

## Goals / Non-Goals

**Goals:**
- Fix the `[] -> null` inconsistency when `tags` is set to `[]` and Kibana's read-back returns a
  nil or empty tags value.
- Preserve `tags` omitted from configuration (null intent) as `null` when the API returns
  nil/empty.
- Preserve a non-empty `tags` list unchanged.
- Resolve this without changing the `tags` schema or the create/update request mapping, which are
  already correct.

**Non-Goals:**
- Adding a shared `ListFromAPIStringsPreserveKnownEmpty` helper in `internal/utils/typeutils`. Per
  explicit human direction on this issue, the fix is implemented inline in
  `dashboardPopulateFromAPI`, mirroring how the `description` fix was done inline rather than via a
  new shared abstraction. A shared helper remains a reasonable follow-up if a third `types.List`
  call site needing this pattern appears, but is out of scope here.
- Changing `Set`-based tags handling in `security_exception_item` or `agentbuildertool` — already
  correct.
- Determining definitively whether Kibana 9.5 GET returns `"tags": []` or omits the key for an
  empty tags array — the intent-preserving fix is correct either way (see Open questions).

## Decisions

**Normalization strategy**: intent-preserving, plan-aware check in `dashboardPopulateFromAPI`,
inline, following the same shape as the existing `description` and `time_range.mode` handling in
that same function:

```go
// Map tags: preserve prior known intent (including known-empty []) when the API
// returns a nil or empty tags value, so a practitioner-set tags = [] does not
// collapse to null on read-back (REQ-009).
if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
    m.Tags = typeutils.SliceToListTypeString(ctx, *data.Data.Tags, path.Root("tags"), &diags)
} else if !m.Tags.IsUnknown() {
    // Prior plan/state tags is already null or a known value (including known-empty
    // []); keep it as-is instead of forcing null.
} else {
    m.Tags = types.ListNull(types.StringType)
}
```

`m.Tags` on entry to `dashboardPopulateFromAPI` already carries the prior plan (on post-create/
post-update read-back) or prior state (on refresh) value, the same precondition the `description`
and `time_range.mode` branches rely on. When `m.Tags` is `Unknown` (for example on import, where
there is no prior plan/state value to preserve), the fix falls back to `types.ListNull`, matching
today's behavior for that case.

**Why not reuse `typeutils.SetFromAPIStringsPreserveKnownEmpty`?** That helper operates on
`types.Set`; dashboard `tags` is `types.List`. Per human direction captured on this issue, the fix
stays inline rather than generalizing a new `types.List` variant of that helper in this change.

**Why not change the `tags` schema (e.g. add a default or `UseStateForUnknown`)?** The bug is
isolated to the read-back mapping; the plan already carries the correct known-empty intent. A
schema-level fix would be a larger, unnecessary change for a read-path bug.

## Risks / Trade-offs

- [Low risk] If Kibana omits the `tags` key entirely for an empty array (rather than returning
  `[]`), the fix still preserves practitioner intent correctly, since the check is intent-based
  (prior known value), not dependent on which empty shape Kibana sends.
- [Low risk] The fix is localized to one branch in `dashboardPopulateFromAPI`; no cascading effects
  on other fields or resources.
- [Low risk] Import continues to map nil/empty API `tags` to `null` (no prior intent to preserve),
  consistent with today's behavior for that path.

## Open questions

- Does Kibana 9.5 GET return `"tags": []` or omit `tags` when the dashboard was created with an
  empty array? Non-blocking for the fix itself (the intent-preserving check is correct either way);
  relevant only for acceptance-test design and for confirming which branch is actually exercised in
  CI against a live stack.
