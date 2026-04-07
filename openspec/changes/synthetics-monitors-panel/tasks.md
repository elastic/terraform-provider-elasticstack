# Tasks: Synthetics Monitors Panel Support

## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [x] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [x] 2.1 Add `synthetics_monitors_config` schema block to `internal/kibana/dashboard/schema.go`
- [x] 2.2 Extend `panelModel` struct in `models_panels.go` with `SyntheticsMonitorsConfig` field
- [x] 2.3 Create `models_synthetics_monitors_panel.go` with read and write converter functions
- [x] 2.4 Determine whether to share filter model types and converters with `synthetics_stats_overview` (REQ-033); if so, extract to a shared file (e.g. `models_synthetics_filters.go`) or reference the existing shared implementation
- [x] 2.5 Update the panel write-path dispatcher in `models_panels.go` to handle `synthetics_monitors` type via the typed config block
- [x] 2.6 Update the panel read-path dispatcher in `models_panels.go` to populate `synthetics_monitors_config` on read-back
- [x] 2.7 Implement null-preservation on read-back: treat empty or absent `config`, `filters`, and individual filter dimension arrays as equivalent to null in state
- [x] 2.8 Add schema validation that `synthetics_monitors_config` is only valid with `type = "synthetics_monitors"` (REQ-006 extension)
- [x] 2.9 Update `config_json` write-path error message in `models_panels.go` to explicitly name `synthetics_monitors` as unsupported (REQ-010 update)
- [x] 2.10 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [x] 3.1 Add acceptance tests for `synthetics_monitors` panel creation with no config block (bare panel)
- [x] 3.2 Add acceptance tests for `synthetics_monitors` panel creation with a `filters` block containing selected filter dimensions
- [x] 3.3 Add acceptance tests for `synthetics_monitors` panel with all five filter dimensions set
- [x] 3.4 Add acceptance tests for plan stability: after create, a subsequent plan SHALL show no changes (no spurious diffs from empty filters)
- [x] 3.5 Add unit tests for the `synthetics_monitors` panel write converter (Terraform model to API payload)
- [x] 3.6 Add unit tests for the `synthetics_monitors` panel read converter (API payload to Terraform model), including empty-filters null-preservation
- [x] 3.7 Verify that setting `config_json` on a panel with `type = "synthetics_monitors"` returns a plan-time or apply-time error diagnostic
