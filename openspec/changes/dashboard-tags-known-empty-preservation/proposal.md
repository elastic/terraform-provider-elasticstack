## Why

Setting `tags = []` on an `elasticstack_kibana_dashboard` resource and applying fails with:

```
Error: provider produced an unexpected new value: .tags: was cty.ListValEmpty(cty.String), but now null.
```

The `tags` attribute is an `Optional` `schema.ListAttribute` of strings
(`internal/kibana/dashboard/schema.go:157-161`). A practitioner who sets `tags = []` plans a
known, empty list. The read-back mapping in `dashboardPopulateFromAPI`
(`internal/kibana/dashboard/models.go:104-109`) collapses any nil-or-empty API `tags` value to
`types.ListNull(types.StringType)`:

```go
if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
    m.Tags = typeutils.SliceToListTypeString(ctx, *data.Data.Tags, path.Root("tags"), &diags)
} else {
    m.Tags = types.ListNull(types.StringType)
}
```

This discards the practitioner's known-empty intent and reports a `[] -> null` inconsistency
regardless of whether Kibana echoes back `"tags": []` or omits the key. The create/update request
mapping already sends an empty array correctly (`models.go:178-183`, `222-227`); the bug is
isolated to this read-back branch. This is the same symptom class already fixed for the dashboard
root `description` (REQ-009, archived change `2026-07-02-dashboard-description-null-normalization`)
and already handled correctly for `tags` on sibling resources that use `types.Set`
(`security_exception_item`, `agentbuildertool`, via
`typeutils.SetFromAPIStringsPreserveKnownEmpty`).

## What Changes

Make the dashboard `tags` read-back intent-preserving: when the API returns a nil or empty tags
value, preserve the prior plan/state `tags` value (including a known-empty `[]`) instead of
forcing `types.ListNull`. Only overwrite `tags` in state when the API returns a non-empty array.
The fix is implemented inline in `dashboardPopulateFromAPI`, following the same pattern already
used for `description` and `time_range.mode` in that function — no new shared helper is added in
this change.

No schema changes are needed. No migration is required.

## Capabilities

### Modified Capabilities

- `kibana-dashboard`: extend REQ-009 (state preservation for fields Kibana omits or defaults) to
  cover the root-level `tags` attribute — when the API returns a nil or empty `tags` value and
  prior plan/state had `tags` as a known value (including known-empty `[]`), the provider SHALL
  preserve that prior value in state rather than forcing `null`.

## Impact

- `internal/kibana/dashboard/models.go` — change the `tags` mapping in `dashboardPopulateFromAPI`
  (around lines 104-109) from unconditional `types.ListNull` on nil/empty API data to an
  intent-preserving check against the prior `m.Tags` value.
- `openspec/specs/kibana-dashboard/spec.md` — delta spec extending REQ-009 to cover root-level
  `tags` normalization.
- Acceptance/unit tests covering: `tags = []` round-trips as `[]` (not `null`) on apply and
  subsequent plan; `tags` omitted from configuration (null intent) stays `null` when the API
  returns nil/empty; a non-empty `tags` list round-trips unchanged.
