## Context

The Kibana Dashboard API defines both `options_list_control` and `range_slider_control` Config objects as a union type with two branches:

- **Field branch** (`KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField` / `KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField`): sources values from a Kibana data view field. Requires `data_view_id` + `field_name`. Has an optional `values_source` that can only legally be `"field"` (default for legacy controls).
- **ES|QL branch** (`KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql` / `KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql`): sources values from an ES|QL query. Requires `esql_query` + `values_source = "esql_query"`.

Both controls share the same structural pattern: two branches with branch-specific discriminating fields and a set of shared behavioral attributes.

Relevant Go types in `generated/kbapi/kibana.gen.go`:

- `KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl_Config` (~line 47333) — union interface.
- `KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql` (~line 46250).
- `KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField` (~line 46333).
- `KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql` (~line 46509).
- `KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField` (~line 46534).

## Goals / Non-Goals

**Goals:**

- Expose the ES|QL variant for both `options_list_control` and `range_slider_control` through typed schema attributes, eliminating the need for `config_json` fallback.
- Restructure both config blocks symmetrically into `by_field {}` / `by_esql {}` branches.
- Ship a Plugin Framework state upgrader (v0 → v1) so existing state migrates automatically.
- Follow the union pattern established by `slo_overview_config` (single/groups), `vis_config` (by_value/by_reference), and `discover_session_config` (by_value/by_reference).

**Non-Goals:**

- Changes to `time_slider_control` or `esql_control` — these are single-shape panels, unaffected.
- Cross-referencing `data_view_id` against a real data view resource — treated as an opaque string per existing resource convention.
- Adding new control types.

## Decisions

### D1: Schema shape — two nested branch blocks

Both config blocks restructure to expose `by_field {}` and `by_esql {}` as sibling `SingleNestedAttribute` blocks. Shared attributes (`title`, `use_global_filters`, `ignore_validations`, `sort`, `display_settings`, etc.) live **inside each branch**, not hoisted to the config root. This:

- Mirrors the API layout (the discriminating and shared fields are all inside the branch object).
- Matches the existing union patterns on this resource.
- Keeps branch schemas symmetric and independently documented.

### D2: `values_source` on `by_field` — hidden, auto-set to `"field"`

The API's Field branch `values_source` accepts only `"field"` (its default for legacy controls). Exposing a schema attribute with a single legal value provides no user value. The model layer sets `values_source = "field"` unconditionally when writing the Field branch. This is an internal implementation detail, not a user-configurable attribute.

On `by_esql`, `values_source` is exposed and **required** — it is a real discriminator (`"esql_query"`) and is required by the API.

### D3: State migration — PF v0 → v1 upgrader

Because the flat attribute layout is replaced by nested branch blocks, the schema version increments from 0 to 1. A `ResourceWithUpgradeState` upgrader rewrites v0 state by moving all current `options_list_control_config` and `range_slider_control_config` flat attributes under a `by_field {}` object. On first `terraform apply` after upgrading the provider, Terraform runs the upgrader automatically. Users then update their HCL to use `by_field { ... }` and run `terraform plan` to confirm no drift.

### D4: `sort` and other shared attributes live inside each branch

The issue author explicitly confirmed this in the human direction comment. The API places `sort { by, direction }` inside the branch objects, not at the control-config root. Keeping shared attrs inside each branch is consistent with `vis_config` and `discover_session_config`.

### D5: Null-preservation on `by_esql` optional boolean attrs

The same REQ-009 null-preservation pattern used on the Field branch applies to ES|QL branches for optional boolean attributes (`use_global_filters`, `ignore_validations`, `exclude`, `exists_selected`, `run_past_timeout` on options_list; `use_global_filters`, `ignore_validations` on range_slider). Only attributes explicitly set by the user are updated from the API response.

### D6: `by_esql.values_source` validation

`values_source` on `by_esql` is required and SHOULD be validated at plan time to accept only `"esql_query"`. This prevents accidental misuse and produces a clear error message.

## Risks / Trade-offs

- [Breaking change] The flat-schema restructure is a breaking change for existing `options_list_control_config` and `range_slider_control_config` users. Mitigated by: the resource is in technical preview; the state upgrader handles automatic state migration; CHANGELOG guidance directs users to wrap their existing flat attributes in `by_field { ... }`.
- [Symmetry overhead] Shared attributes duplicated inside each branch increase schema code volume. Accepted: this matches the existing pattern and keeps branches independent.
- [State upgrader complexity] Writing the v0 → v1 upgrader requires careful attribute-by-attribute mapping. Mitigated by test coverage of the upgrade path.

## Open questions

None — all design decisions were resolved in the human direction comment by @tobio on 2026-07-01.
