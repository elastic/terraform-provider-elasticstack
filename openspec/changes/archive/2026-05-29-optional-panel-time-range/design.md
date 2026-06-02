## Context

Every typed Lens chart panel write goes through `lenscommon.LensChartPresentationWritesFor`, which calls `resolver.ResolveChartTimeRange(chartLevel)` and unconditionally assigns the result to `writes.TimeRange`. The `Resolver` interface returns a value type (`kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema`), making it impossible to signal "omit this field". The `ResolveChartTimeRange` implementation falls back to the dashboard-level time range and ultimately to `now-15m`/`now`, so a non-empty struct is always produced and always sent.

For by-reference vis panels, `LensByReferenceAttributes()` marks `time_range` as `Required: true`, and `VisByReferenceModelToAPIConfig1` always dereferences `byRef.TimeRange` without a nil check. Both the API type (`KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1.TimeRange`) and the by-value chart types (`KibanaHTTPAPIsXyChartNoESQL.TimeRange`, etc.) declare the field as `*Type, omitempty`, confirming the API has always been willing to accept absence.

## Goals / Non-Goals

**Goals:**
- Omit `time_range` from the Kibana API payload when the practitioner has not configured it on the panel.
- Make `vis_config.by_reference.time_range` optional in the Terraform schema.
- Update the `Resolver` interface and `ResolveChartTimeRange` so callers get a pointer that is nil when no chart-level time range is configured.
- Remove the dashboard-inheritance fallback and the hardcoded `now-15m`/`now` fallback.
- Update the read path to drop the now-unnecessary dashboard-comparable null-preservation check.

**Non-Goals:**
- Changing dashboard-level `time_range` (remains required).
- Changing raw `config_json` panel behavior.
- Introducing a migration path for practitioners who relied on implicit inheritance (they must now set `time_range` explicitly on each panel if desired).

## Decisions

### 1. Change `Resolver.ResolveChartTimeRange` to return a pointer

**Decision**: Change the return type from `kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema` to `*kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema`.

**Rationale**: The existing value-type return forced callers to always have a non-nil result. A pointer return lets `nil` mean "omit from API payload". The `LensChartPresentationWrites.TimeRange` field is already `*Type`, so callers (`chart.TimeRange = writes.TimeRange`) require no changes — nil flows through naturally and `omitempty` handles the rest.

**Alternative considered**: Keep value return, add a separate boolean flag. Rejected — pointer is idiomatic Go for "optional value".

### 2. `ResolveChartTimeRange` returns nil when chart-level is unset

**Decision**: When `chartLevel == nil`, return `nil`. Remove dashboard inheritance and the hardcoded fallback entirely.

**Rationale**: The inheritance was a workaround for a perceived API requirement that never existed. Removing it makes the write path honest: "send exactly what the practitioner configured".

### 3. `VisByReferenceModel.TimeRange` becomes a pointer

**Decision**: Change `TimeRange VisByReferenceTimeRangeModel` to `TimeRange *VisByReferenceTimeRangeModel` in `models/lens.go`.

**Rationale**: Terraform Plugin Framework represents an absent `Optional` single-nested attribute as a nil pointer in the Go model. Making the field a pointer is the standard PF pattern for optional nested attributes.

**Alternative considered**: Keep the value type and use a separate `TimeRangeSet bool` field. Rejected — PF pointer convention is simpler and consistent with the rest of the codebase.

### 4. Remove dashboard-comparable null-preservation from the read path

**Decision**: Remove the `DashboardLensComparableTimeRange` check inside `chartTimeRangeFromAPI`.

**Rationale**: That check existed to prevent drift when the API echoed back the dashboard time range (which we had written). Once we stop writing the dashboard time range, Kibana will not echo it back for panels that don't have one configured. The check becomes dead code. The remaining nil/empty guard is still correct and sufficient.

**Alternative considered**: Keep the check as a safety net. Rejected — dead code obscures intent and the nil guard already handles the case.

### 5. `HasLensByReferenceShapeAtRoot` detects by `ref_id` only

**Decision**: Remove the `time_range` presence check from `HasLensByReferenceShapeAtRoot`; detect by-reference shape solely by the presence of a non-empty `ref_id` string.

**Rationale**: `time_range` is no longer a defining characteristic of the by-reference wire format. `ref_id` is the only required field (`json:"ref_id"` without omitempty), so it remains a reliable discriminator.

## Risks / Trade-offs

[Behavioral breaking change for implicit inheritance users] → No Terraform state migration needed, but practitioners relying on implicit dashboard time range inheritance will see panels lose their custom time range on next apply. This is the correct behavior — the mitigation is clear release notes and a changelog entry.

[`HasLensByReferenceShapeAtRoot` false positives] → A raw `config_json` blob that happens to contain `ref_id` but is not a by-reference vis panel could now be misdetected. In practice, `ref_id` is a Kibana by-reference convention that does not appear in by-value payloads, so the risk is low. If needed, a follow-up can add a secondary discriminator.

[`VisByReferenceModel` model pointer change] → Any code that dereferences `byRef.TimeRange` without a nil check will panic. The write path in `VisByReferenceModelToAPIConfig1` and the read path in `PopulateVisByReferenceTFModelFromAPIConfig1` must both be updated to guard against nil. These are the only two call sites.

## Migration Plan

No Terraform state migration is required:
- Making `Required → Optional` for `by_reference.time_range` is backward-compatible; existing state with a value continues to work.
- By-value panels that had an implicitly-inherited time range will have it omitted on the next apply; Kibana will then return no panel-level time range on the following read, producing a stable null state that matches the practitioner's config.

A changelog entry in `CHANGELOG.md` under **Bug Fixes** should describe the behavioral change and note that practitioners who want an explicit panel-level time range must now set it in config.
