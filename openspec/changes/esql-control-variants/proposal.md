## Why

The `options_list_control` and `range_slider_control` dashboard panel types support two source variants in the Kibana API: a **Field** variant (data view + field name) and an **ES|QL** variant (an arbitrary ES|QL query). The Terraform resource currently exposes only the Field variant for each. Users who need ES|QL-backed controls must fall back to `config_json`, losing typed validation and drift-safe planning.

## What Changes

Both `options_list_control_config` and `range_slider_control_config` are restructured from a flat schema to a **two-branch union** pattern:

- `by_field {}` — the existing Field variant. All current flat attributes are relocated inside this nested block. `data_view_id` and `field_name` are required within the block.
- `by_esql {}` — the new ES|QL variant. `esql_query` and `values_source` are required within this block. Shared attributes (`title`, `use_global_filters`, etc.) are present in both branches.

Panel-level validators enforce exactly one of the two branches. `values_source` on `by_field` is not exposed as a Terraform attribute; the model layer sets it to `"field"` automatically.

Because the resource carries a technical-preview marker, a breaking schema restructure is acceptable. A Plugin Framework `ResourceWithUpgradeState` (v0 → v1) state upgrader ships alongside so existing state files are migrated automatically on the first `terraform apply`. Users update their HCL after the upgrade.

## Capabilities

### New Capabilities

None (ES|QL control is an existing API feature now exposed in the typed schema).

### Modified Capabilities

- `kibana-dashboard`: restructure REQ-027 (`options_list_control_config`) and REQ-028 (`range_slider_control_config`) to introduce the `by_field` / `by_esql` union and add ES|QL branch support. Add a new REQ covering state migration.

## Impact

- `internal/kibana/dashboard/panel/optionslist/` — schema, model, API converter restructured to the two-branch union.
- `internal/kibana/dashboard/panel/rangeslider/` — same restructure.
- `internal/kibana/dashboard/` — `resource.go` wires up the `ResourceWithUpgradeState` upgrader; schema version bumped from 0 to 1.
- Unit tests updated for both control panels; state-upgrade round-trip tests added.
- Acceptance tests updated: one per control type demonstrating both `by_field` and `by_esql` configurations.
- Pinned-panels support (`pinned_panels` block) inherits the restructured schemas automatically.
