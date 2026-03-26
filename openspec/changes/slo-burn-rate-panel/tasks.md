# Tasks: SLO Burn Rate Panel Support

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [ ] 2.1 Add `slo_burn_rate_config` schema block to `internal/kibana/dashboard/schema.go`
- [ ] 2.2 Extend `panelModel` struct in `models_panels.go` with `SloBurnRateConfig` field
- [ ] 2.3 Create `models_slo_burn_rate_panel.go` with read and write converter functions
- [ ] 2.4 Update the panel write-path dispatcher in `models_panels.go` to handle `slo_burn_rate` type via the typed config block
- [ ] 2.5 Update the panel read-path dispatcher in `models_panels.go` to populate `slo_burn_rate_config` on read-back
- [ ] 2.6 Add schema validation that `slo_burn_rate_config` is only valid with `type = "slo_burn_rate"` (REQ-006 extension)
- [ ] 2.7 Add schema validator for `duration` using regex `^\d+[mhd]$`
- [ ] 2.8 Implement null-preservation for `slo_instance_id`: when prior state is null and API returns `"*"`, keep null in state
- [ ] 2.9 Update `config_json` write-path error message in `models_panels.go` to explicitly name `slo_burn_rate` as unsupported (REQ-010 update)
- [ ] 2.10 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [ ] 3.1 Add acceptance tests for `slo_burn_rate` panel creation with required fields (`slo_id` and `duration`)
- [ ] 3.2 Add acceptance tests for `slo_burn_rate` panel with `slo_instance_id` set to a specific instance
- [ ] 3.3 Add acceptance tests for `slo_burn_rate` panel with a `drilldowns` list
- [ ] 3.4 Add acceptance tests for plan-time validation rejection of an invalid `duration` value (e.g. `"5x"`)
- [ ] 3.5 Add acceptance tests confirming `slo_instance_id` is null in state when not configured, even after Kibana read-back
- [ ] 3.6 Add unit tests for the `slo_burn_rate` panel write converter (Terraform model to API payload)
- [ ] 3.7 Add unit tests for the `slo_burn_rate` panel read converter (API payload to Terraform model)
- [ ] 3.8 Verify that setting `config_json` on a panel with `type = "slo_burn_rate"` returns a plan-time or apply-time error diagnostic
