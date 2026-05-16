# Design: Migrate snapshot resources to entitycore envelope

## Overview

Move both snapshot-related Plugin SDK resources to Plugin Framework and the entitycore envelope. These are independent but share the snapshot domain.

## 1. Snapshot Lifecycle (SLM) — `elasticstack_elasticsearch_snapshot_lifecycle`

### Current State
- File: `internal/elasticsearch/cluster/slm.go` (~360 LOC)
- SDK resource with cron schedule, indices list, retention object, and various snapshot flags.

### Schema Mapping
| SDK Field | PF Type |
|-----------|---------|
| `name` | `StringAttribute`, Required, RequiresReplace |
| `schedule` | `StringAttribute`, Required |
| `repository` | `StringAttribute`, Required |
| `config` | `SingleNestedAttribute` or object block |
| `retention` | `SingleNestedAttribute` |
| `indices` | `ListAttribute` of string |
| `ignore_unavailable`, `include_global_state`, etc. | `BoolAttribute` |

### Callback Design
- **Create/Update**: Build `models.SlmPolicy` → `PutSlm` → set ID → return
- **Read**: `GetSlm` → populate model
- **Delete**: `DeleteSlm`

### Model
Flatten the nested `config` and `retention` objects into the top-level model or use nested structs with `tfsdk` tags.

## 2. Snapshot Repository — `elasticstack_elasticsearch_snapshot_repository`

### Current State
- File: `internal/elasticsearch/cluster/snapshot_repository.go` (~540 LOC)
- SDK resource with polymorphic per-type sub-schemas: `fs`, `url`, `azure`, `gcs`, `s3`
- Uses ` ExactlyOneOf` validation across repository type blocks.

### Schema Mapping
The polymorphic design in SDK uses multiple `TypeList` blocks, exactly one of which must be set. In PF, this maps to:
- Multiple `SingleNestedAttribute` blocks (one per type)
- Custom `ValidateConfig` or plan-time validation ensuring exactly one is set
- Or use a single `StringAttribute` (`type`) + `SingleNestedAttribute` (`settings`) dynamic object

**Decision**: Keep the per-type block design (closest to existing interface), using `SingleNestedAttribute` for each repository type. Add `ValidateConfig` on the concrete resource type to enforce exactly-one.

### Common Sub-Schema
Fields like `chunk_size`, `compress`, `max_snapshot_bytes_per_sec`, `max_restore_bytes_per_sec`, `readonly` appear in all types. In PF, define a reusable object type or embed in each type block.

### Callback Design
- **Create/Update**: Build `models.SnapshotRepository` based on which type block is set → `PutSnapshotRepository` → set ID
- **Read**: `GetSnapshotRepository` → determine type from response → populate correct nested block
- **Delete**: `DeleteSnapshotRepository`

### Model
```go
type Data struct {
    ID                      types.String `tfsdk:"id"`
    Name                    types.String `tfsdk:"name"`
    Fs                      types.Object `tfsdk:"fs"`
    Url                     types.Object `tfsdk:"url"`
    Azure                   types.Object `tfsdk:"azure"`
    Gcs                     types.Object `tfsdk:"gcs"`
    S3                      types.Object `tfsdk:"s3"`
    ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}
```

## Registration
- Remove both from `provider/provider.go` `ResourcesMap`
- Add both to `provider/plugin_framework.go` `resources()`

## Testing
- `slm_test.go` and `snapshot_repository_test.go` acceptance tests should pass after migration.
- Repository type switching (e.g., fs → s3) should ForceNew (name is already ForceNew, so full replace).
