# Tasks: Pie Chart Structured Legend

## 1. Spec

- [x] 1.1 Keep the delta spec aligned with `proposal.md` and `design.md`
- [x] 1.2 On completion, sync the delta into the canonical dashboard spec or archive the change

## 2. Schema

- [x] 2.1 Replace `pie_chart_config.legend_json` with an optional `pie_chart_config.legend` nested block in `internal/kibana/dashboard/schema.go`
- [x] 2.2 Reuse `getPartitionLegendSchema()` for the pie chart legend block so pie, treemap, and mosaic share the same Terraform-facing legend shape
- [x] 2.3 Keep the pie chart legend block optional in the schema
- [x] 2.4 Update pie chart descriptions so the schema documents `visible` rather than raw API `visibility`

## 3. Models and Helpers

- [x] 3.1 Replace `LegendJSON` in `pieChartConfigModel` with `Legend *partitionLegendModel`
- [x] 3.2 Add pie-specific partition legend helper methods in `internal/kibana/dashboard/models_partition_helpers.go` for `kbapi.PieLegend`
- [x] 3.3 Remove pie legend JSON normalization/marshal helpers that are no longer needed after the typed legend conversion

## 4. Converters

- [x] 4.1 Update pie chart read converters to populate `pie_chart_config.legend` from the API `PieLegend` object
- [x] 4.2 Update pie chart write converters to build the API `PieLegend` object from the typed `legend` block
- [x] 4.3 Preserve current omitted-legend write behavior by sending `size = "auto"` when the optional `legend` block is absent
- [x] 4.4 Ensure the typed pie legend uses Terraform field `visible` while mapping to the API `visibility` field

## 5. Testing

- [x] 5.1 Update pie chart acceptance test fixtures to replace `legend_json` with the structured `legend` block
- [x] 5.2 Update pie chart acceptance assertions to validate nested `legend.size`, `legend.visible`, and any other exercised legend fields
- [x] 5.3 Add or update unit tests for pie chart API-to-model round-trip using typed legend conversion
- [x] 5.4 Add coverage for the omitted-legend case so the optional block behavior is verified explicitly
- [x] 5.5 Remove tests that assert or ignore `pie_chart_config.legend_json`

## 6. Documentation and Requirement Alignment

- [x] 6.1 Update OpenSpec REQ-023 and the schema summary for `pie_chart_config`
- [x] 6.2 Review any pie chart examples or test data that still use raw legend JSON and convert them to the structured block
