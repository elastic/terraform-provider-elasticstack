# Tasks: Synthetics Monitors Panel Support

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [ ] 2.1 Add `synthetics_monitors_config` schema block to `internal/kibana/dashboard/schema.go`
- [ ] 2.2 Extend `panelModel` struct in `models_panels.go` with `SyntheticsMonitorsConfig` field
- [ ] 2.3 Create `models_synthetics_monitors_panel.go` with read and write converter functions
- [ ] 2.4 Determine whether to share filter model types and converters with `synthetics_stats_overview` (REQ-033); if so, extract to a shared file (e.g. `models_synthetics_filters.go`) or reference the existing shared implementation
- [ ] 2.5 Update the panel write-path dispatcher in `models_panels.go` to handle `synthetics_monitors` type via the typed config block
- [ ] 2.6 Update the panel read-path dispatcher in `models_panels.go` to populate `synthetics_monitors_config` on read-back
- [ ] 2.7 Implement null-preservation on read-back: treat empty or absent `config`, `filters`, and individual filter dimension arrays as equivalent to null in state
- [ ] 2.8 Add schema validation that `synthetics_monitors_config` is only valid with `type = "synthetics_monitors"` (REQ-006 extension)
- [ ] 2.9 Update `config_json` write-path error message in `models_panels.go` to explicitly name `synthetics_monitors` as unsupported (REQ-010 update)
- [ ] 2.10 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [ ] 3.1 Add acceptance tests for `synthetics_monitors` panel creation with no config block (bare panel)
- [ ] 3.2 Add acceptance tests for `synthetics_monitors` panel creation with a `filters` block containing selected filter dimensions
- [ ] 3.3 Add acceptance tests for `synthetics_monitors` panel with all six filter dimensions set
- [ ] 3.4 Add acceptance tests for plan stability: after create, a subsequent plan SHALL show no changes (no spurious diffs from empty filters)
- [ ] 3.5 Add unit tests for the `synthetics_monitors` panel write converter (Terraform model to API payload)
- [ ] 3.6 Add unit tests for the `synthetics_monitors` panel read converter (API payload to Terraform model), including empty-filters null-preservation
- [ ] 3.7 Verify that setting `config_json` on a panel with `type = "synthetics_monitors"` returns a plan-time or apply-time error diagnostic
