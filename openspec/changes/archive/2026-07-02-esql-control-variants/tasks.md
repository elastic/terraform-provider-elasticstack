## 1. Schema restructure — options_list_control

- [x] 1.1 In `internal/kibana/dashboard/panel/optionslist/schema.go`, replace the flat `Attributes` map with two `SingleNestedAttribute` blocks: `by_field` and `by_esql`. Add the `ExactlyOneOf` panel-level validator (following the `slooverview` pattern using `validators.ExactlyOneOfNestedAttrsValidator`).
- [x] 1.2 `by_field` block: move `data_view_id` (required), `field_name` (required), `title`, `use_global_filters`, `ignore_validations`, `single_select`, `exclude`, `exists_selected`, `run_past_timeout`, `search_technique`, `selected_options`, `display_settings`, `sort` into the nested attribute.
- [x] 1.3 `by_esql` block: add `esql_query` (required, string), `values_source` (required, string, validator: `OneOf("esql_query")`), `title`, `use_global_filters`, `ignore_validations`, `single_select`, `exclude`, `exists_selected`, `run_past_timeout`, `search_technique`, `selected_options`, `display_settings`, `sort` (same shared attrs as `by_field`).
- [x] 1.4 Add `objectvalidator.ConflictsWith` on each branch block to enforce mutual exclusion.

## 2. Schema restructure — range_slider_control

- [x] 2.1 In `internal/kibana/dashboard/panel/rangeslider/schema.go`, apply the same two-branch restructure: `by_field` and `by_esql`, with `ExactlyOneOf` validator.
- [x] 2.2 `by_field` block: move `data_view_id` (required), `field_name` (required), `title`, `use_global_filters`, `ignore_validations`, `value`, `step` into the nested attribute.
- [x] 2.3 `by_esql` block: add `esql_query` (required), `values_source` (required, validator: `OneOf("esql_query")`), `title`, `use_global_filters`, `ignore_validations`, `value`, `step`.

## 3. Model updates — options_list_control

- [x] 3.1 In `internal/kibana/dashboard/panel/optionslist/model.go`, redefine `OptionsListControlConfigModel` to hold `ByField` and `ByEsql` typed sub-models instead of flat attributes. (Model structs live in `internal/kibana/dashboard/models/controls.go`; the optionslist package's `model.go` holds the `PopulateFromAPI`/`BuildConfig` conversion logic.)
- [x] 3.2 Define `ByFieldModel` with all Field-branch attributes (including `DataViewId`, `FieldName`, etc.). (Named `OptionsListControlByFieldModel`.)
- [x] 3.3 Define `ByEsqlModel` with all ES|QL-branch attributes (including `EsqlQuery`, `ValuesSource`, plus shared). (Named `OptionsListControlByEsqlModel`.)
- [x] 3.4 Update `model_test.go` to cover both branch shapes.

## 4. Model updates — range_slider_control

- [x] 4.1 In `internal/kibana/dashboard/panel/rangeslider/model.go`, apply the same model restructure.
- [x] 4.2 Define `ByFieldModel` (`DataViewId`, `FieldName`, shared) and `ByEsqlModel` (`EsqlQuery`, `ValuesSource`, shared).
- [x] 4.3 Update `model_test.go`.

## 5. API converter — options_list_control

- [x] 5.1 In `internal/kibana/dashboard/panel/optionslist/api.go`, update `ToAPI`: check which branch is non-null; when `by_field`, build a `KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField` struct with `ValuesSource = "field"` (not exposed to user); when `by_esql`, build `KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql` with `ValuesSource` from the model. (`BuildConfig`/`buildFieldConfig`/`buildEsqlConfig` live in `model.go`; `api.go`'s `Handler.ToAPI` is unchanged and delegates to `BuildConfig`.)
- [x] 5.2 Update `FromAPI`: inspect the returned config discriminant (via raw JSON key presence: `esql_query` only exists on the ES|QL branch); populate `ByField` or `ByEsql` in the model accordingly, applying null-preservation semantics for optional boolean attrs on both branches.
- [x] 5.3 Update `api_test.go` with both branch round-trip tests.

## 6. API converter — range_slider_control

- [x] 6.1 Same dual-branch ToAPI / FromAPI update for `internal/kibana/dashboard/panel/rangeslider/api.go`.
- [x] 6.2 Update `api_test.go`.

## 7. State upgrader (v0 → v1)

- [x] 7.1 Bump the dashboard resource schema version from 0 to 1 in `internal/kibana/dashboard/resource.go` (or wherever schema version is declared). (Version set in `internal/kibana/dashboard/schema.go`'s `getSchema()`.)
- [x] 7.2 Implement a `ResourceWithUpgradeState` upgrader for version 0 → 1: for each panel entry whose `type` is `"options_list_control"`, relocate the flat `options_list_control_config` attributes under a `by_field {}` object; do the same for `"range_slider_control"` panels. Pinned-panels entries are included. (Also covers `sections[].panels[]`, which share the same panel envelope.)
- [x] 7.3 Write a state upgrade test for each of the two control types verifying that a v0 flat-attribute state is correctly rewritten to v1 branch format.

## 8. Tests

- [x] 8.1 Unit tests: options_list Field branch round-trip (model ↔ API), null-preservation, validator reject of missing/both branches.
- [x] 8.2 Unit tests: options_list ES|QL branch round-trip, `values_source` validator.
- [x] 8.3 Unit tests: range_slider Field branch and ES|QL branch round-trips.
- [x] 8.4 State upgrade tests (task 7.3 above).
- [x] 8.5 Acceptance test for `options_list_control` demonstrating `by_field` and (in a separate step) `by_esql` config.
- [x] 8.6 Acceptance test for `range_slider_control` demonstrating both branches.
- [x] 8.7 Acceptance test verifying that a pre-upgrade (v0 flat) state is successfully migrated on plan/apply.

## 9. Documentation and CHANGELOG

- [x] 9.1 Add a **Breaking change** CHANGELOG entry explaining the restructure and the migration path (wrap existing attributes in `by_field { ... }`; run `terraform apply` to let the state upgrader run automatically). (This repo's CHANGELOG.md is auto-generated from the PR's `## Changelog` section per `dev-docs/high-level/contributing.md`; the breaking-change note is included in the PR description, not a manual CHANGELOG.md edit.)
- [x] 9.2 Update any provider documentation examples for `options_list_control` and `range_slider_control` to use the new `by_field {}` / `by_esql {}` syntax. (Updated `examples/guides/guide2-operations/main.tf` and regenerated `docs/` via `make docs-generate`.)

## 10. Spec sync and validation

- [x] 10.1 Run `make check-openspec` and resolve any failures. (230 passed, 0 failed.)
- [x] 10.2 Run `make build` and `go vet ./...`. (Both clean; `make build`'s golangci-lint step reports 0 issues.)
- [x] 10.3 Run `go test ./internal/kibana/dashboard/...` (non-acceptance unit tests). (All packages pass.)
