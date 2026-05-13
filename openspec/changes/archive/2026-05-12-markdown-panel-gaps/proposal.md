## Why

The Kibana Dashboard API's `kbn-dashboard-panel-type-markdown` schema models markdown panels as a union of "by value" (inline `content` with required `settings.open_links_in_new_tab`, optional `description`, `hide_title`, `title`, `hide_border`) and "by reference" (linked library item via `ref_id` plus the same presentation fields). The Terraform `markdown_config` block exposes only `content`, `description`, `hide_title`, and `title` — it cannot set `open_links_in_new_tab`, cannot set `hide_border`, and cannot reference a markdown library item at all. Practitioners must drop down to `config_json` to use either shape, defeating the typed-block experience.

## What Changes

- Add a required nested `settings = object({ open_links_in_new_tab = bool })` block to `markdown_config` for the by-value branch (matches the API's required field).
- Add `hide_border = bool` to `markdown_config`.
- Restructure `markdown_config` to expose the two API branches symmetrically: `by_value = object({ content, settings, description, hide_title, title, hide_border })` and `by_reference = object({ ref_id, description, hide_title, title, hide_border })`. Exactly one of `by_value` / `by_reference` SHALL be set.
- **BREAKING (pre-release)**: existing flat `markdown_config` shape (`content`, `description`, `hide_title`, `title`) becomes nested under `by_value`. No migration support; cut over.
- Continue to support `config_json` for `markdown` panels (existing REQ-010 behavior).

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: extend REQ-012 (Markdown panel behavior) to cover the by-value/by-reference union, the required `settings.open_links_in_new_tab` field, and `hide_border`.

## Impact

- `internal/kibana/dashboard/schema.go` — restructure `markdown_config` schema.
- `internal/kibana/dashboard/models_markdown_panel.go` — extend the model and read/write helpers; map the union via the generated kbapi `KbnDashboardPanelTypeMarkdownConfig0` / `KbnDashboardPanelTypeMarkdownConfig1` types.
- Validators for "exactly one of by_value / by_reference set".
- Update existing markdown unit tests and the markdown acceptance test to the new shape.
- Add acceptance coverage for the by-reference branch (requires creating a markdown library item beforehand).
- Update the example under `examples/resources/elasticstack_kibana_dashboard/` if it uses `markdown_config`.
