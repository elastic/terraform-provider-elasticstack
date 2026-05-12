## 1. Schema restructure

- [ ] 1.1 Replace flat `markdown_config` with `by_value` / `by_reference` sub-blocks per the design
- [ ] 1.2 Add `settings = object({ open_links_in_new_tab = bool })` (required) under `by_value`
- [ ] 1.3 Add `hide_border` to both sub-blocks
- [ ] 1.4 Conditional validators: exactly one of `by_value` / `by_reference` set
- [ ] 1.5 Update descriptions under `internal/kibana/dashboard/descriptions/`

## 2. Model and mapping

- [ ] 2.1 Restructure the markdown panel model in `models_markdown_panel.go` to mirror the new schema
- [ ] 2.2 Map to `KbnDashboardPanelTypeMarkdownConfig0` (by-value) and `KbnDashboardPanelTypeMarkdownConfig1` (by-reference) in write
- [ ] 2.3 Detect the API branch on read and populate the matching sub-block, leaving the other null
- [ ] 2.4 Apply REQ-009 null-preservation to `open_links_in_new_tab` and `hide_border`

## 3. Tests

- [ ] 3.1 Update existing unit tests in `models_markdown_panel_test.go` for the new shape
- [ ] 3.2 Add unit tests for the by-reference branch and the discriminator validators
- [ ] 3.3 Update the existing markdown acceptance test to the new shape; add a by-reference acceptance test that creates a markdown library item via the saved-objects API beforehand
- [ ] 3.4 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...` (`TF_ACC=1` for acceptance)

## 4. Examples

- [ ] 4.1 Update any markdown panel examples under `examples/resources/elasticstack_kibana_dashboard/` to the new shape

## 5. Spec sync

- [ ] 5.1 Run `make check-openspec`
