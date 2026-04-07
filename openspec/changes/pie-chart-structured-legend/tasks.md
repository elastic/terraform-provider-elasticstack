# Tasks: Pie Chart Structured Legend

## 1. Spec

- [ ] 1.1 Keep the delta spec aligned with `proposal.md` and `design.md`
- [ ] 1.2 On completion, sync the delta into the canonical dashboard spec or archive the change

## 2. Schema

- [ ] 2.1 Replace `pie_chart_config.legend_json` with an optional `pie_chart_config.legend` nested block in `internal/kibana/dashboard/schema.go`
- [ ] 2.2 Reuse `getPartitionLegendSchema()` for the pie chart legend block so pie, treemap, and mosaic share the same Terraform-facing legend shape
- [ ] 2.3 Keep the pie chart legend block optional in the schema
- [ ] 2.4 Update pie chart descriptions so the schema documents `visible` rather than raw API `visibility`

## 3. Models and Helpers

- [ ] 3.1 Replace `LegendJSON` in `pieChartConfigModel` with `Legend *partitionLegendModel`
- [ ] 3.2 Add pie-specific partition legend helper methods in `internal/kibana/dashboard/models_partition_helpers.go` for `kbapi.PieLegend`
- [ ] 3.3 Remove pie legend JSON normalization/marshal helpers that are no longer needed after the typed legend conversion

## 4. Converters

- [ ] 4.1 Update pie chart read converters to populate `pie_chart_config.legend` from the API `PieLegend` object
- [ ] 4.2 Update pie chart write converters to build the API `PieLegend` object from the typed `legend` block
- [ ] 4.3 Preserve current omitted-legend write behavior by sending `size = "auto"` when the optional `legend` block is absent
- [ ] 4.4 Ensure the typed pie legend uses Terraform field `visible` while mapping to the API `visibility` field

## 5. Testing

- [ ] 5.1 Update pie chart acceptance test fixtures to replace `legend_json` with the structured `legend` block
- [ ] 5.2 Update pie chart acceptance assertions to validate nested `legend.size`, `legend.visible`, and any other exercised legend fields
- [ ] 5.3 Add or update unit tests for pie chart API-to-model round-trip using typed legend conversion
- [ ] 5.4 Add coverage for the omitted-legend case so the optional block behavior is verified explicitly
- [ ] 5.5 Remove tests that assert or ignore `pie_chart_config.legend_json`

## 6. Documentation and Requirement Alignment

- [ ] 6.1 Update OpenSpec REQ-023 and the schema summary for `pie_chart_config`
- [ ] 6.2 Review any pie chart examples or test data that still use raw legend JSON and convert them to the structured block
