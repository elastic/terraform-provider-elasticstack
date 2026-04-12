# Proposal: Pie Chart Structured Legend

## Why

The dashboard resource currently exposes pie chart legend configuration through `pie_chart_config.legend_json`, a normalized JSON string that mirrors the raw Kibana API object. Other partition-style charts in the same resource, such as treemap and mosaic, expose legend settings through a typed Terraform object with first-class attributes. That inconsistency makes pie charts harder to author, harder to discover in schema docs, and harder to keep aligned with the rest of the dashboard resource.

The API shape for pie chart legends already matches the partition legend model used by treemap and mosaic: the legend contains `nested`, `size`, `truncate_after_lines`, and `visibility`. This means the provider can expose the same richer Terraform shape for pie charts without inventing a new data model. The only intentional difference is that pie legend remains optional in the schema, preserving the current pie-chart contract rather than adopting treemap/mosaic requiredness.

This is a breaking schema change because the public Terraform field changes from `legend_json` to `legend`, and the nested attribute name changes from raw API `visibility` to provider-style `visible`. That break is acceptable here because the dashboard resource is unreleased and no compatibility migration is required.

## What Changes

- **Replace `pie_chart_config.legend_json` with `pie_chart_config.legend`** as an optional structured nested block.
- **Reuse the existing partition legend shape** for pie charts: `nested`, `size`, `truncate_after_lines`, and `visible`.
- **Keep pie legend optional** in the Terraform schema, unlike treemap and mosaic.
- **Update REQ-023 and the schema summary** in the dashboard OpenSpec spec so pie charts use `dataset_json` plus structured `legend`, rather than `dataset_json` plus raw `legend_json`.
- **Update pie chart models, converters, and tests** to use typed legend mapping instead of JSON marshaling/unmarshaling for legend values.

## Capabilities

After this change, practitioners will be able to:

- Configure pie chart legends using a typed Terraform block instead of authoring raw JSON.
- Use the same legend attribute names and enum values as the other partition chart blocks in this resource.
- Avoid the raw API-only field name `visibility`; pie charts will use the Terraform attribute `visible` like treemap and mosaic.
- Continue omitting the pie chart legend block entirely when they do not need to override legend settings.

## Impact

- **Breaking schema change**: `pie_chart_config.legend_json` is removed and replaced by `pie_chart_config.legend`.
- **No migration plan required**: the resource is unreleased, so compatibility shims and state upgraders are unnecessary.
- **Optionality preserved**: pie chart `legend` remains optional even though it shares the partition legend shape.
- **Converter change only for legend**: pie chart dataset, metrics, group-by, and mode-selection behavior are otherwise unchanged.
- **Testing impact**: pie chart acceptance tests and unit tests must be updated to use the structured legend block and validate round-trip behavior.
