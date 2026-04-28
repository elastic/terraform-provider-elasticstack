## 1. Schema and validation alignment

- [ ] 1.1 Update `internal/kibana/slo/schema.go` to add `filter_kql`, `good_kql`, and `total_kql` object-form attributes under `kql_custom_indicator`
- [ ] 1.2 Add schema-level exclusivity and conditional validation so each KQL field accepts either the legacy string form or the new `_kql` form, but not both
- [ ] 1.3 Add provider-side conditional validators for indicator-specific required and forbidden fields using `internal/utils/validators/conditional.go`
- [ ] 1.4 Align simple validators for `slo_id`, custom metric aggregations, metric names, and `time_window.type`
- [ ] 1.5 Add `settings.sync_field`, `artifacts`, and `enabled` to the SLO Terraform schema

## 2. Model and client behavior

- [ ] 2.1 Extend the SLO Terraform model types in `internal/kibana/slo/models.go` and `models_kql_custom_indicator.go` to serialize and read back both KQL union forms
- [ ] 2.2 Extend settings and artifact mapping so `sync_field` and `artifacts.dashboards[].id` flow through create, update, and read
- [ ] 2.3 Add `enabled` to the internal SLO model and read-path mapping from Kibana responses
- [ ] 2.4 Extend `internal/clients/kibanaoapi/slo.go` with helper functions for enable and disable operations and wire them into SLO create and update reconciliation
- [ ] 2.5 Verify whether additive schema changes require a state upgrader and implement one if needed

## 3. Verification and documentation

- [ ] 3.1 Add or update unit tests for KQL union conversion, settings mapping, artifact mapping, enabled handling, and tightened validation
- [ ] 3.2 Add or update acceptance tests covering `_kql` inputs, `enabled` behavior, `sync_field`, and validation failures for invalid SLO configurations
- [ ] 3.3 Regenerate `docs/resources/kibana_slo.md` and confirm the new `_kql`, `sync_field`, `artifacts`, and `enabled` fields are documented clearly
- [ ] 3.4 Run the relevant build and test commands for the touched SLO code paths and resolve any regressions
