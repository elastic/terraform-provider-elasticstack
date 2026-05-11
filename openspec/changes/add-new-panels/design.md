## Context

Three panel types in the recently bumped Kibana Dashboard API spec lack typed Terraform support: `image` (small, ~190 spec lines), `slo_alerts` (small, ~85 spec lines), and `discover_session` (large, ~580 spec lines). All three are bundled because:

1. Two of them are small enough that splitting them across changes adds review and refactor overhead without commensurate benefit.
2. All three need a `url_drilldown` nested block and (for two of them) a `time_range` nested block. Extracting these once and reusing them avoids three near-identical schema copies.
3. The drilldown and time-range shapes are also needed by the upcoming per-Lens-chart presentation work; extracting them here pays the refactor cost once.

The resource is not yet released, so breaking schema changes inside this bundle are acceptable; no migration path is required.

## Goals / Non-Goals

**Goals:**

- Practitioners can author all three panel types entirely from HCL, consistent with existing typed panels.
- Plan-time validators catch shape errors (discriminator violations, enum mismatches, length bounds) before round-tripping to Kibana.
- Round-trip cleanly with REQ-009 null-preservation for optional fields and drilldown API defaults.
- The shared `url_drilldown` and `time_range` blocks become reusable building blocks for follow-on changes (per-Lens-chart presentation, vis_by_reference).

**Non-Goals (v1):**

- Managing uploaded image files (`image_config.src.file.file_id` references a Kibana file id; file upload is out of scope and will be addressed by a future `elasticstack_kibana_file` resource if requested).
- Typing the inner Discover session tab `data_source` shape field-by-field — both DSL and ES|QL data sources are handled via `data_source_json` for v1.
- Surfacing `references` on `discover_session_config.by_reference` — deferred pending confirmation of whether the Dashboard API expects client-side references for this panel or resolves them from `ref_id` internally.
- Drilldown variants the API does not allow for the relevant panel today (e.g., `dashboard_drilldown` on `slo_alerts` / `discover_session`). The schema is shaped so adding them later is additive.
- Multi-tab Discover panels — the API restricts `tabs` to a single entry today and we expose a single `tab` object accordingly.

## Decisions

### Cross-cutting

- **Bundle scope**: all three panel types ship together with shared `url_drilldown` and `time_range` schema helpers. Splitting them by panel was rejected because two of the three are small and the shared helpers are simpler to extract once than to retrofit.
- **Drilldowns are typed everywhere**: `slo_alerts` and `discover_session` use the shared `url_drilldown` block. `image` uses a discriminated union with `dashboard_drilldown` xor `url_drilldown` because it is the only panel that supports both today. `drilldowns_json` was rejected for these panels — the surface is small, well-bounded, and benefits from plan-diff visibility. *(See `expose-lens-chart-presentation-fields` for the parallel decision on Lens chart blocks.)*
- **Shared `url_drilldown` schema** lives in a small helper (e.g., `schema_shared_drilldowns.go`) and is consumed by the new panels plus the existing `slo_burn_rate_config` and `slo_overview_config` drilldowns. This is a Go-level consolidation — user-visible behavior of existing SLO panels does not change.
- **Shared `time_range` schema** mirrors the existing dashboard-root shape (`from`, `to`, `mode?`). The dashboard-root attribute keeps its current schema; the shared helper is extracted so panel-level `time_range` (here, and in follow-on changes) uses the identical shape.
- **`config_json` is forbidden for all three panel types**, consistent with the typed-only stance for `options_list_control`, `synthetics_*`, `esql_control`, and `lens-dashboard-app`.

### `image_config`

- **`src` as nested discriminated sub-blocks**, not a flat `src_type`/`file_id`/`url` triple. Aligns with the API and gives plan-time mutex validation.
  - *Rejected*: flat fields with conflict validators — gives less obvious error messages and forces a `src_type` enum that duplicates the discriminator already present in the API.
- **`drilldowns` as a typed discriminated union** (per-entry: `dashboard_drilldown` xor `url_drilldown`). Image is unique in supporting both. The `dashboard_drilldown` sub-block fields mirror the API (`dashboard_id`, `label`, `trigger = "on_click_image"`, `use_filters`, `use_time_range`, `open_in_new_tab`).
- **`object_fit` default not baked in**: API defaults to `"contain"`. Leave the TF attribute optional with no plan-side default so REQ-009 null-preservation keeps unset state null after read.
- **`file_id` lifecycle is out of scope**: the resource accepts the id as a string only. A future `elasticstack_kibana_file` resource (or data source) can manage the uploaded asset. Documented in resource docs and called out in the example.

### `slo_alerts_config`

- **`slos` is required with `len(slos) > 0`**. The API allows an empty list (default `[]`) but a zero-SLO alerts panel is a misconfiguration — surface the error at plan time rather than letting the user create a useless panel.
- **`slo_instance_id` null-preservation** mirrors the existing `slo_burn_rate_config` behavior: when Kibana returns the server-side default `"*"` and prior state had it null, state stays null.
- **`drilldowns` reuses the shared `url_drilldown` block**. Only `on_open_panel_menu` is allowed by the API today; validator constrains the trigger.

### `discover_session_config` — Option C (typed envelope + targeted JSON escape hatches)

- **`by_value` xor `by_reference`** mirrors `lens_dashboard_app_config`'s discriminator. Exactly one set; conditional validator returns a diagnostic when both or neither are present.
- **Single `tab` object** (not `tabs = list`). The API restricts `tabs` to one entry. A list-of-one in TF would force `tab[0]` indexing for no current benefit; `tabs = [...]` can be added later as a non-breaking superset if Kibana lifts the cardinality limit. The single-`tab` choice matches authoring ergonomics today.
- **`tab.dsl` xor `tab.esql`** mirrors `datatable_config`'s `no_esql`/`esql` precedent. Lets the same outer `tab` host both variants without prematurely flattening the schema.
- **`data_source_json` (not typed sub-blocks)** for both DSL and ES|QL `data_source` unions. Consistent with how Lens layer data sources are handled today (`data_source_json` on `datatable_config` / `mosaic_config` / `treemap_config` etc.) and avoids re-implementing data-view authoring inside the dashboard resource. *Rejected*: typed `data_view_reference`/`data_view_spec` sub-blocks — high surface, high churn, low marginal value for plan-diff quality.
- **`filters = list(object({ filter_json = string }))`** in the DSL tab, matching the `dashboard-filters` design exactly. Practitioners learn one filter shape across the resource.
- **`header_row_height` and `row_height` are typed as strings with validators** matching the API union (`"1".."5"|"auto"` for header, `"1".."20"|"auto"` for row). Established Terraform pattern for `number | "auto"` API unions; avoids a forked attribute pair.
- **`time_range` is optional at the panel level**, even though the API marks it required on both branches. Practitioners SHALL be able to omit it to inherit the dashboard-root `time_range`; the model layer applies the dashboard value when the panel attribute is null at write time. Read preserves null per REQ-009 when the API echoes the inherited value back. *(See "Risks" for why this is safe.)*
- **`selected_tab_id` is optional input and computed when omitted**, since practitioners typically don't know tab UUIDs.
- **`references` is deferred**: not exposed on `by_reference` in v1. The decision pivots on whether the Dashboard API expects client-side `references` for this panel — to be confirmed against a stack experimentally; if needed, a follow-on change adds `references_json` mirroring `lens_dashboard_app_config.by_reference.references_json`.
- **`drilldowns` reuses the shared `url_drilldown` block** (URL-only with `on_open_panel_menu` trigger).

### Coupling and sequencing

- **`dashboard-filters` should land before this change** so the DSL tab `filters` shape can adopt the same `list(object({ filter_json = string }))` schema element it produces. If `dashboard-filters` slips, this change ships with the same shape independently; the two designs are deliberately identical to avoid drift.
- **`fix-dashboard-unknown-panel-preservation` should land before this change** so there is no window where a previously-unknown panel becomes typed mid-flight. Soft dependency: correctness holds either way.

## Risks / Trade-offs

- [Risk] Practitioners may want to type Discover session inner tab state (column overrides, ES|QL data sources) field-by-field in the future, which would be a breaking change for anyone relying on `data_source_json` ➝ *Mitigation*: the resource is not yet released; we are free to promote subsets of JSON to typed attributes in subsequent changes. Document the v1 boundary in resource docs.
- [Risk] API may extend drilldowns for `slo_alerts` or `discover_session` to include `dashboard_drilldown` ➝ *Mitigation*: the typed block list is shaped so adding a new sub-block per entry is additive (the per-entry shape becomes a discriminator just like image's).
- [Risk] Optional panel-level `time_range` on `discover_session` differs from the API's "required" marking ➝ *Mitigation*: the model layer materializes the dashboard-root `time_range` into the panel payload at write time when the panel attribute is null, so the API always sees a value. On read, null-preservation keeps state intent stable.
- [Risk] `references` deferral could turn out to be a hard requirement at runtime (panel write fails without it) ➝ *Mitigation*: planned spike before implementation begins — verify behavior against a live Kibana with a Discover session saved object. If references are required, the follow-on `references_json` addition is additive (no schema break) and can land before this change is unblocked for archive.
- [Risk] `slos` non-empty validation rejects configurations the API itself accepts ➝ *Mitigation*: this is a deliberate opinionated constraint; surfaced clearly in resource docs. Empty SLO alerts panels are not a meaningful authoring outcome.
- [Risk] Shared `url_drilldown` extraction subtly changes existing SLO panel behavior ➝ *Mitigation*: the Go refactor is mechanical (extract-method), unit tests for `slo_burn_rate` / `slo_overview` already cover drilldown round-trip, and acceptance tests run as part of the bundle's verification.
- [Risk] Acceptance test for `by_reference` Discover panels needs a Discover session saved object as a fixture ➝ *Mitigation*: use the saved-objects API in test setup, mirroring the lens-by-reference acceptance pattern already proven in `acc_lens_dashboard_app_panels_test.go`.

## Open questions to confirm during implementation

- Empirically verify whether `discover_session_config.by_reference` write requires the dashboard request to include client-side `references` for the panel. If yes, fold a `references_json` attribute into this change (additive) before archival.
- Confirm Kibana behavior when `discover_session_config.by_value` is created without an explicit `time_range`: does the API accept it inheriting the dashboard time range, or does it require a payload value? The dashboard-time-range fallback at write time addresses the latter case; the former simply works.
