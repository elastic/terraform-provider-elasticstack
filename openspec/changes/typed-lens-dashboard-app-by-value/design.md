## Context

The dashboard resource currently supports two Lens-related panel paths:

- `type = "vis"` Lens panels, which expose rich typed Terraform chart blocks such as `xy_chart_config`, `metric_chart_config`, `pie_chart_config`, and related chart types.
- `type = "lens-dashboard-app"` panels, which expose `lens_dashboard_app_config` with `by_reference` typed fields and `by_value.config_json` as an opaque JSON object.

The generated Kibana API client models by-value `lens-dashboard-app` config as `KbnDashboardPanelTypeLensDashboardAppConfig0`. The generated `vis` Lens inline config is `KbnDashboardPanelTypeVisConfig0`. These two generated unions wrap the same chart structs for the chart types already supported by the provider, even though the enclosing dashboard panel discriminators differ.

The completed `lens-dashboard-app-panel` change intentionally scoped typed by-value blocks out. This change extends that work without changing by-reference behavior or the existing `vis` panel contract.

## Goals / Non-Goals

**Goals:**

- Allow practitioners to author `lens-dashboard-app` by-value panels using typed Lens chart blocks instead of only opaque `config_json`.
- Reuse existing chart schema/model/converter code where practical.
- Keep `by_value.config_json` available for unsupported chart shapes, future Kibana fields, and exact API payload authoring.
- Preserve the practitioner-selected representation on read: typed input should stay typed when the API response can be mapped, and raw JSON input should stay raw.
- Keep by-reference `lens-dashboard-app` behavior unchanged.
- Keep `type = "vis"` Lens panel behavior unchanged.

**Non-Goals:**

- Migrating existing `vis` Lens panels to `lens-dashboard-app`.
- Removing or deprecating `by_value.config_json`.
- Modeling every generated Lens field that is not already exposed by the existing typed chart blocks.
- Adding typed drilldown or saved-object reference modeling for by-value charts in this change.
- Changing generated `kbapi` code.

## Decisions

### Add typed chart blocks inside `lens_dashboard_app_config.by_value`

`by_value` will contain the existing raw `config_json` plus typed chart blocks for supported Lens chart types. Exactly one by-value source must be configured.

Example shape:

```hcl
lens_dashboard_app_config = {
  by_value = {
    metric_chart_config = {
      title            = "Requests"
      data_source_json = jsonencode(...)
      query = {
        language = "kql"
        query    = ""
      }
      metrics = [...]
    }
  }
}
```

Rationale: nesting the typed chart source under `by_value` keeps the existing `by_value` / `by_reference` mode split intact and avoids making top-level `xy_chart_config`, `metric_chart_config`, and similar blocks valid for two panel discriminators with different wire shapes.

Alternative considered: allow existing top-level typed chart blocks when `type = "lens-dashboard-app"`. That would reduce schema duplication but would weaken the current top-level mutual-exclusion model and make it less obvious whether the panel is using the `vis` or `lens-dashboard-app` API discriminator.

### Reuse existing chart converters through a union adapter

The existing typed chart converters build and read `KbnDashboardPanelTypeVisConfig0`. The by-value `lens-dashboard-app` API expects `KbnDashboardPanelTypeLensDashboardAppConfig0`. Because both unions serialize the same chart structs, conversion can bridge through JSON:

1. Build a scratch `panelModel` containing the selected typed chart model.
2. Use the existing `lensVisualizationConverter.buildAttributes` method to produce `KbnDashboardPanelTypeVisConfig0`.
3. Marshal that union to JSON.
4. Unmarshal the JSON into `KbnDashboardPanelTypeLensDashboardAppConfig0`.
5. Assign that value to `KbnDashboardPanelTypeLensDashboardApp.Config`.

Read-back can reverse the bridge when prior Terraform state or plan selected a typed `by_value` chart:

1. Classify the raw `lens-dashboard-app.config` as by-value using the existing root `type` discriminator rule.
2. Convert the raw config JSON into `KbnDashboardPanelTypeVisConfig0`.
3. Detect chart type with the existing `detectLensVizType`.
4. Use the existing converter's `populateFromAttributes` to populate the typed model.
5. Fall back to `config_json` if the response cannot be represented by the selected typed chart block.

Rationale: this avoids duplicating per-chart conversion logic and keeps typed chart behavior aligned between `vis` and `lens-dashboard-app` wherever the generated chart structs are shared.

Alternative considered: create a second full converter interface for `KbnDashboardPanelTypeLensDashboardAppConfig0`. That would be explicit but would duplicate most conversion logic and increase the risk that `vis` and `lens-dashboard-app` typed chart behavior diverge.

### Preserve raw JSON and representation intent

`by_value.config_json` remains supported and semantically normalized. The provider should not convert raw JSON-authored or imported by-value panels into typed chart blocks unless prior Terraform configuration selected that typed chart block.

Rationale: raw JSON is the highest-fidelity representation. Auto-converting imported panels into typed blocks could drop unsupported fields or create surprising state shape changes.

Alternative considered: always prefer typed chart blocks on read when the chart type is recognized. That would make imports friendlier for supported charts but risks losing fields the existing typed chart models do not expose.

### Do not model by-value references and drilldowns yet

The generated by-value chart structs include fields such as `references`, `drilldowns`, `hide_title`, `hide_border`, and `time_range`. Existing typed `vis` chart blocks do not expose every one of those fields uniformly. This change should first reuse the established chart schema surface; practitioners who need exact payload control can continue using `by_value.config_json`.

Rationale: adding common by-value reference/drilldown/display wrappers would be useful, but it is a separate design problem because those fields must be merged into many chart structs without creating destructive read/write behavior.

Alternative considered: add `references_json`, `drilldowns_json`, `hide_title`, `hide_border`, and `time_range` directly under `by_value` and merge them into typed charts. This can be added later once the typed chart adapter is proven.

## Risks / Trade-offs

- Typed chart blocks may not expose every by-value Lens API field → Keep `config_json` as the escape hatch and preserve raw JSON representation by default.
- JSON bridge failures could surface late during conversion → Add targeted unit tests for each supported typed chart adapter path and diagnostics that name the selected by-value chart source.
- Existing chart converters inject defaults such as `lensPanelTimeRange()` → Treat this as acceptable only where the generated by-value chart schema requires the same field; verify with tests that the produced `lens-dashboard-app.config` is valid and stable.
- Some generated chart structs may exist in both unions but behave differently in Kibana when used under `lens-dashboard-app` → Start with chart types already accepted by the current `lens-dashboard-app` generated union and add acceptance coverage for representative chart families.
- Schema duplication can drift between top-level `vis` chart blocks and nested `by_value` chart blocks → Reuse schema helper functions instead of copying attribute maps by hand where possible.

## Migration Plan

This change is additive. Existing dashboards using `lens_dashboard_app_config.by_value.config_json`, `lens_dashboard_app_config.by_reference`, or `type = "vis"` typed chart blocks remain valid.

Rollback is straightforward: configurations using the new typed by-value blocks can be rewritten to equivalent `by_value.config_json` payloads. No state migration is expected.

## Open Questions

1. Should every current typed `vis` Lens chart block be exposed immediately under `by_value`, or should implementation start with a narrower set such as metric, XY, pie, and waffle?
2. Should a follow-up change add common by-value fields such as `references_json`, `drilldowns_json`, `hide_title`, `hide_border`, and explicit `time_range` wrappers?
3. Are there Kibana versions where a chart struct accepted in `KbnDashboardPanelTypeVisConfig0` is generated into `KbnDashboardPanelTypeLensDashboardAppConfig0` but rejected by the dashboard API at runtime?
