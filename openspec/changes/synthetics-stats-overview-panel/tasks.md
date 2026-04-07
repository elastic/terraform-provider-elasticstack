# Tasks: Synthetics Stats Overview Panel Support

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [x] 2.1 Add `synthetics_stats_overview_config` schema block to `internal/kibana/dashboard/schema.go`
- [x] 2.2 Extend `panelModel` struct in `models_panels.go` with `SyntheticsStatsOverviewConfig` field
- [x] 2.3 Create `models_synthetics_stats_overview_panel.go` with read and write converter functions
- [x] 2.4 Update the panel write-path dispatcher in `models_panels.go` to handle `synthetics_stats_overview` type via the typed config block
- [x] 2.5 Update the panel read-path dispatcher in `models_panels.go` to populate `synthetics_stats_overview_config` on read-back
- [x] 2.6 Add schema validation that `synthetics_stats_overview_config` is only valid with `type = "synthetics_stats_overview"` (REQ-006 extension)
- [x] 2.7 Update `config_json` write-path error message in `models_panels.go` to explicitly name `synthetics_stats_overview` as unsupported (REQ-010 update)
- [x] 2.8 Implement read-back null preservation: empty or absent API config maps to null `synthetics_stats_overview_config` in state
- [x] 2.9 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [x] 3.1 Add acceptance tests for `synthetics_stats_overview` panel creation with no config (empty panel showing all monitors)
- [x] 3.2 Add acceptance tests for `synthetics_stats_overview` panel with display settings (`title`, `description`, `hide_title`, `hide_border`)
- [x] 3.3 Add acceptance tests for `synthetics_stats_overview` panel with `filters` block (at least one filter category)
- [x] 3.4 Add acceptance tests for `synthetics_stats_overview` panel with `drilldowns`
- [x] 3.5 Add unit tests for the write converter (Terraform model to API payload)
- [x] 3.6 Add unit tests for the read converter (API payload to Terraform model), including the empty-config-to-null case
- [x] 3.7 Verify that setting `config_json` on a panel with `type = "synthetics_stats_overview"` returns a plan-time or apply-time error diagnostic
