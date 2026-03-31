## 1. Spec

- [x] 1.1 Validate the change with `./node_modules/.bin/openspec validate slo-overview-panel`.
- [ ] 1.2 Sync or archive the delta into `openspec/specs/kibana-dashboard/spec.md` after implementation is verified.

## 2. Schema

- [x] 2.1 Add `slo_overview_config` block to `internal/kibana/dashboard/schema.go` with `single` and `groups` nested blocks and all documented attributes.
- [x] 2.2 Add schema-level validators to enforce that `slo_overview_config` is only present on `type = "slo_overview"` panels, that exactly one of `single` or `groups` is set, and that `single.slo_id` is required when the `single` block is configured.
- [x] 2.3 Add enum validator for `groups.group_filters.group_by` restricted to `"slo.tags"`, `"status"`, `"slo.indicator.type"`, and `"_index"`.
- [x] 2.4 Extend the `panelModel` struct in `internal/kibana/dashboard/models_panels.go` to carry the new `slo_overview_config` field and route it through the panel dispatcher.

## 3. Converter

- [x] 3.1 Create `internal/kibana/dashboard/models_slo_overview_panel.go` implementing write conversion from Terraform state to the Kibana API payload, selecting `slo-single-overview-embeddable` or `slo-group-overview-embeddable` based on the active sub-block.
- [x] 3.2 Implement read conversion from the Kibana API payload back to Terraform state, including null preservation for `slo_instance_id` when not configured.
- [x] 3.3 Implement drilldowns round-trip mapping, with normalization for any Kibana-injected defaults identified during development.
- [x] 3.4 Implement `filters_json` normalization for `groups.group_filters.filters_json` using the same JSON-semantic-equality approach used by other `*_json` attributes.

## 4. Tests

- [x] 4.1 Add acceptance tests in `internal/kibana/dashboard/acc_test.go` covering single-mode SLO overview panel lifecycle (create, plan-clean, import, destroy).
- [x] 4.2 Add acceptance tests covering groups-mode SLO overview panel lifecycle, including `group_filters` with `group_by` and `kql_query`.
- [x] 4.3 Add a schema validation test confirming that configuration with both `single` and `groups` blocks is rejected.
- [x] 4.4 Add unit tests in `models_slo_overview_panel_test.go` covering write and read conversion for both modes.
- [x] 4.5 Add a unit test covering null preservation for `slo_instance_id` when Kibana returns `"*"` and prior state is null.

## 5. Verification

- [x] 5.1 Run `make build` and verify no compilation errors.
- [x] 5.2 Run targeted Go tests for `internal/kibana/dashboard` and any touched packages.
- [ ] 5.3 Run acceptance tests for `internal/kibana/dashboard` against a live stack using the environment described in `dev-docs/high-level/testing.md`.
