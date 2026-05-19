## Why

Kibana's Data Views update API preserves unspecified fields (documented as "Only the specified properties are updated in the data view. Unspecified fields stay as they are persisted."). When a Terraform configuration omits `data_view.runtime_field_map` after previously defining it, Kibana keeps the runtime fields. The provider then reads them back into state, but because `runtime_field_map` is `Optional` and not `Computed`, Terraform rejects the post-apply state with:

```
produced an unexpected new value: .data_view.runtime_field_map: was null, but now ...
```

This prevents clean updates when users remove `runtime_field_map` from configuration (reported in #2135). Adding `Computed: true` to the schema attribute aligns Terraform's expectations with Kibana's partial-update semantics.

## What Changes

- **Schema change**: Add `Computed: true` to `data_view.runtime_field_map` in `internal/kibana/dataview/schema.go`
- **Acceptance test fix**: Update `TestAccResourceDataView` so the `basic → basic_updated` transition exercises a true in-place update rather than a resource replacement:
  - Keep `field_attrs` in `basic_updated` testdata to avoid `RequiresReplace`
  - Add `captureID` / `checkIDUnchanged` assertions across update steps
  - Update expectations for `runtime_field_map` in the update step to assert the fields are still present (because Kibana preserves them)
  - Add `ImportStateVerifyIgnore` for `data_view.runtime_field_map` in the import step
- **Requirements update**: Update REQ-011 in `kibana-data-view` spec to cover the preserved-null behavior for `runtime_field_map` when the attribute is omitted from config

## Capabilities

### New Capabilities
*(none)*

### Modified Capabilities

- `kibana-data-view`: REQ-011 (State mapping for empty collections) must be updated to specify that `runtime_field_map` is `Computed`. When the attribute is omitted from config and Kibana preserves a non-empty `runtime_field_map`, the provider SHALL accept the API response into state without raising a consistency error.

## Impact

- `internal/kibana/dataview/schema.go` — add `Computed: true` to `runtime_field_map`
- `internal/kibana/dataview/acc_test.go` — add ID stability checks, adjust runtime_field_map assertions, add import ignore
- `internal/kibana/dataview/testdata/TestAccResourceDataView/basic_updated/data_view.tf` — keep `field_attrs` to avoid replacement
- `openspec/specs/kibana-data-view/spec.md` — update REQ-011 requirements
