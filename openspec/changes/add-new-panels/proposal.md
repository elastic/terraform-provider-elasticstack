## Why

The recent Kibana Dashboard API spec bump exposes three panel types that the Terraform resource does not currently model: `image`, `slo_alerts`, and `discover_session`. Until the `fix-dashboard-unknown-panel-preservation` change lands, dashboards containing these panels are silently corrupted on read; even after that fix lands they can only be round-tripped opaquely and cannot be authored from HCL. This change adds typed support for all three so practitioners can create, read, update, and delete dashboards that contain them.

The three panels are bundled into a single change because two of them are small, all three share two reusable building blocks (a `url_drilldown` schema and a `time_range` schema), and decoupling them would force the same design conversation three times and pay a refactor cost twice.

## What Changes

Bundle: three new typed panel blocks on `panels[]`, plus two shared schema helpers extracted from the bundle.

### Shared schema helpers (implementation prep)

- Extract a shared `url_drilldown` nested-block schema used by `slo_alerts_config.drilldowns`, `discover_session_config.drilldowns`, the new image panel's drilldown union, and the existing `slo_burn_rate_config.drilldowns` / `slo_overview_config.drilldowns` (no behavior change for existing panels — Go-level consolidation only).
- Extract a shared `time_range` nested-block schema (`object({ from = string, to = string, mode? = string })`) used by `discover_session_config` and made available to future per-Lens-chart presentation work. The existing dashboard-root `time_range` keeps its current shape and behavior unchanged.

### `image_config` block (panel `type = "image"`)

- `src` as mutually exclusive nested sub-blocks: `file = object({ file_id = string })` xor `url = object({ url = string })`.
- `alt_text`, `object_fit` (enum `fill`/`contain`/`cover`/`none`, no TF default), `background_color`.
- `title`, `description`, `hide_title`, `hide_border`.
- `drilldowns` as a typed list with **two** discriminated sub-blocks per entry: `dashboard_drilldown` xor `url_drilldown`. Image is the only panel that supports `dashboard_drilldown` (`on_click_image` trigger).
- Forbid `config_json` for `image` panels.

### `slo_alerts_config` block (panel `type = "slo_alerts"`)

- `slos = list(object({ slo_id = string, slo_instance_id = optional string }))`, **required**, with a `len(slos) > 0` validator (the API allows empty; we reject it to prevent useless panels).
- `title`, `description`, `hide_title`, `hide_border`.
- `drilldowns` as a list of the shared `url_drilldown` block (this panel only supports `on_open_panel_menu` URL drilldowns).
- `slo_instance_id` null-preservation mirroring `slo_burn_rate_config`.
- Forbid `config_json` for `slo_alerts` panels.

### `discover_session_config` block (panel `type = "discover_session"`)

Option C hybrid (typed envelope + targeted JSON escape hatches):

- `title`, `description`, `hide_title`, `hide_border`.
- `by_value` xor `by_reference` (mirrors `lens_dashboard_app_config`).
- **`by_value`**: optional shared `time_range`; **single `tab` object** today (a future `tabs = list(...)` is a non-breaking superset when the API lifts its current `maxItems = 1`). The tab has two mutually exclusive sub-blocks `dsl` xor `esql` (mirrors `datatable_config`'s `no_esql`/`esql` precedent):
  - `dsl`: `column_order`, `column_settings`, `sort`, `density`, `header_row_height` (string with validator `"1".."5"|"auto"`), `row_height` (string with validator `"1".."20"|"auto"`), `rows_per_page`, `sample_size`, `view_mode` (enum), typed `query` block (existing shape), `data_source_json` (JSON escape hatch for the `data_view_reference` / `data_view_spec` union), `filters = list(object({ filter_json = string }))` (consistent with the `dashboard-filters` shape).
  - `esql`: `column_order`, `column_settings`, `sort`, `density`, `header_row_height`, `row_height`, `data_source_json` (ES|QL data source).
- **`by_reference`**: optional shared `time_range`; required `ref_id`; optional `selected_tab_id` (optional input + computed); typed `overrides` block (8 simple scalars).
- `drilldowns` as a list of the shared `url_drilldown` block (this panel only supports `on_open_panel_menu`).
- **`references` deliberately omitted from `by_reference` for v1** — defer until we confirm whether the Dashboard API expects client-side references on this panel or resolves them internally from `ref_id`.
- Forbid `config_json` for `discover_session` panels.

### Cross-cutting

- All three panel blocks conflict with each other, with all other typed panel config blocks, and with `config_json`.
- All three apply REQ-009 null-preservation to optional fields (including drilldown API defaults like `encode_url`, `open_in_new_tab`, `use_filters`, `use_time_range`).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: adds three new panel-behavior requirements — **REQ-040** (image / `image_config`), **REQ-041** (`slo_alerts` / `slo_alerts_config`), **REQ-042** (`discover_session` / `discover_session_config`). REQ-010 (`config_json` typed-only enumeration and panel-type rejection rules) is extended to list `image`, `slo_alerts`, and `discover_session` alongside existing typed-only panel types; each new REQ also declares `config_json` rejected for its panel `type`.

## Impact

### Source files

- `internal/kibana/dashboard/schema.go` — new schema blocks (or split into per-panel files in line with existing large panels: `schema_image_panel.go`, `schema_slo_panel.go` for `slo_alerts`, `schema_discover_session_panel.go`).
- Shared schema helpers: a small `schema_shared_drilldowns.go` (or similar) holding the `url_drilldown` factory and a `schema_time_range.go` for the shared `time_range` block. Existing SLO panels' drilldown schemas refactored to call the shared factory (no behavior change).
- `internal/kibana/dashboard/models_image_panel.go`, `models_slo_alerts_panel.go`, `models_discover_session_panel.go` — new models with read/write helpers.
- `internal/kibana/dashboard/models_panels.go` — wire the three new panel types into `mapPanelFromAPI` and the panel write dispatcher.
- Validators for: `src` discriminator (image), drilldown discriminator (image), `slos` non-empty (slo_alerts), `dsl` xor `esql` (discover), `by_value` xor `by_reference` (discover), `header_row_height` / `row_height` enum-or-bounded-number string (discover), `view_mode` enum (discover), drilldown trigger enums.

### Tests

- Unit tests next to each new model: `models_image_panel_test.go`, `models_slo_alerts_panel_test.go`, `models_discover_session_panel_test.go` covering both src/by-value/by-reference branches, drilldown discriminators, null-preservation, and validator failures.
- Existing SLO panel unit tests adjusted minimally if the shared `url_drilldown` factory changes signatures (target: zero behavioral diff).
- Acceptance tests:
  - `acc_image_panels_test.go` covering both `src` variants and a drilldown of each kind.
  - `acc_slo_alerts_panels_test.go` referencing an SLO created in test setup (reuse fixtures from `acc_slo_*` tests).
  - `acc_discover_session_panels_test.go` covering `by_value` (both `dsl` and `esql` tabs) and `by_reference` (creates a Discover saved object in test setup, mirroring the lens-by-reference acceptance test pattern).

### Examples

- One example per panel type under `examples/resources/elasticstack_kibana_dashboard/`.

### Dependencies and sequencing

- **Soft dependency on `fix-dashboard-unknown-panel-preservation`**: this change is additive on top of the preservation fix; landing the fix first removes a window where typed support and preserved-unknown handling could collide.
- **Soft dependency on `dashboard-filters`**: the Discover session DSL tab's `filters` shape is intentionally identical to dashboard-level `filters` (per `dashboard-filters/design.md`). Landing `dashboard-filters` first lets us reuse the same `filter_json` list element schema; if it slips, this change ships with the same shape independently.
- Out of scope: a separate `elasticstack_kibana_file` resource for managing uploaded image files (`image_config.src.file.file_id` references a file id; file upload is outside this change). Documented as future work.
- Out of scope: typing the inner `tabs[]` ES|QL `data_source` shape (handled via `data_source_json`). Out of scope: `references` on `discover_session_config.by_reference` (deferred pending API confirmation).
- Out of scope: drilldowns on `discover_session` of type `dashboard_drilldown` — the API does not allow them for this panel today.
