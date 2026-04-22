# Design: Pie Chart Structured Legend

## Context

Pie charts are currently the outlier among the dashboard resource's partition-style Lens panels. Treemap and mosaic expose a typed `legend` block backed by the shared `getPartitionLegendSchema()` schema and `partitionLegendModel`, while pie charts expose `legend_json` as a raw normalized JSON string.

At the Kibana API level, pie chart legend shape already matches the same partition legend concept:

- `nested`
- `size`
- `truncate_after_lines`
- `visibility`

The provider therefore does not need a pie-specific schema concept for legend configuration. It only needs to stop treating the pie legend as opaque JSON and instead map it through the same Terraform-facing shape already used for the other partition charts.

The user explicitly wants pie legend optionality preserved. This means pie charts should share the same legend field shape as treemap and mosaic, but not the same requiredness.

## Goals

1. Replace `pie_chart_config.legend_json` with a typed `pie_chart_config.legend` block.
2. Reuse the existing partition legend Terraform shape and validation rules.
3. Keep `pie_chart_config.legend` optional.
4. Preserve current pie-chart behavior when the block is omitted by continuing to send a valid API legend object.
5. Keep the implementation change narrowly scoped to the pie legend surface rather than redesigning other pie chart fields.

## Non-Goals

- Preserving backward compatibility with `legend_json`.
- Changing pie chart `dataset_json`, `metrics`, `group_by`, or query-mode behavior.
- Making pie chart legend required like treemap and mosaic.
- Introducing a new generic partition chart abstraction beyond what already exists.

## Decisions

### Terraform shape: optional `legend` block using the partition legend schema

`pie_chart_config` will replace:

```hcl
legend_json = jsonencode({
  size       = "auto"
  visibility = "visible"
})
```

with:

```hcl
legend = {
  size    = "auto"
  visible = "visible"
}
```

The block uses the existing partition legend fields:

- `nested` - optional bool
- `size` - required string (`auto`, `s`, `m`, `l`, `xl`)
- `truncate_after_lines` - optional number
- `visible` - optional string (`auto`, `visible`, `hidden`)

The field name is intentionally `visible`, not `visibility`, because this change aligns pie charts to the provider's existing partition legend contract rather than the raw API JSON field names.

### Optional block, required `size` when present

The `legend` block itself remains optional. This preserves the current pie chart schema contract.

If the block is present, `size` is required because that is how `getPartitionLegendSchema()` is already defined and because the Kibana API requires a legend size value in the pie legend payload.

### Write-path behavior when legend is omitted

When `pie_chart_config.legend` is omitted, the provider should continue the current behavior of sending a valid pie legend with `size = "auto"` so the API payload remains acceptable.

This keeps the omission path semantically equivalent to today's `legend_json = null` handling, while still allowing explicit typed authoring when the user wants to control legend settings.

### Shared model with pie-specific conversion helpers

The implementation should reuse `partitionLegendModel` for the Terraform representation, but add pie-specific conversion helpers in `models_partition_helpers.go`:

- `fromPieLegend(api kbapi.PieLegend)`
- `toPieLegend() kbapi.PieLegend`

This keeps the pie model aligned with the shared partition legend behavior while respecting the separate generated API types (`kbapi.PieLegend`, `kbapi.TreemapLegend`, `kbapi.MosaicLegend`).

### Read-path behavior

On read-back, the provider should populate `pie_chart_config.legend` from the API legend object using the typed model. The raw `legend_json` field is removed and must no longer appear in state.

The `visible` Terraform attribute maps to the API's `legend.visibility` field.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| The contract is breaking because `legend_json` disappears and nested field names change | Accept the break because the resource is unreleased; document it clearly in the proposal and delta spec |
| The optional block plus required nested `size` may be confusing | Document that the block is optional, but if present it must include `size`; add acceptance coverage for both omitted and populated forms |
| The current omission path may repopulate legend values from the API on refresh | Add explicit tests for the omitted-legend case and confirm whether the provider needs any state-preservation logic beyond the converter change |
| Reusing the partition legend model could hide small pie-specific differences in the future | Limit sharing to the Terraform model shape and keep pie-specific API conversion helpers separate |

## Migration and State

No compatibility migration is planned. The unreleased dashboard resource can take a direct breaking change from `legend_json` to `legend`.

## Open Questions

1. When a user omits `pie_chart_config.legend`, does the current pie panel implementation already round-trip without post-apply or post-refresh drift, or will the new typed shape expose a pre-existing omission issue more clearly?
2. Should the implementation preserve a null legend block in state when Kibana returns the default legend, or is the current "explicit default on refresh" behavior acceptable for this unreleased resource?
