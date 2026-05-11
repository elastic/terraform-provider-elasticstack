## Why

The `elasticstack_elasticsearch_index` resource exposes `sort_field` and `sort_order` for Elasticsearch's `index.sort.*` static settings, but is missing `index.sort.missing` and `index.sort.mode`. All four `index.sort.*` settings are **immutable static settings** that can only be set at index creation time.

Additionally, the current design has structural problems: `sort_field` is modeled as a `SetAttribute`, which does not preserve insertion order. Since sort settings are positionally coupled (field index N corresponds to order/missing/mode index N), using a Set means the provider cannot guarantee correct field→setting mapping for multi-field sorts. Users also have no way to see all per-field sort options in one place.

## What Changes

- **New `sort` attribute**: A `ListNestedAttribute` that groups `field`, `order`, `missing`, and `mode` per sort entry. Fixes the structural coupling issue and delivers all four sort settings.
- **Deprecate `sort_field` and `sort_order`**: Both attributes remain functional but receive `DeprecationMessage` and `ConflictsWith` validators. Users mixing old and new attributes get a plan-time error.
- **`RequiresReplaceIf` with private state**: Store the ordered sort config from Elasticsearch in private state during `Read`. Use this to suppress replace when migrating from legacy attributes to the new `sort` block with semantically equivalent settings (including `missing`/`mode`, with explicit defaults treated the same as absent settings).
- **Add `sort.missing` and `sort.mode` to `staticSettingsKeys`**: So the `use_existing` adoption flow correctly compares these settings against the existing index.
- **Update `use_existing.go`**: Add `case "sort.missing"` and `case "sort.mode"` branches.

## Capabilities

### New Capabilities

- _(none — the `sort` attribute is added to the existing `elasticstack_elasticsearch_index` resource)_

### Modified Capabilities

- **`elasticsearch-index`**: Add `sort` `ListNestedAttribute`; deprecate `sort_field` and `sort_order`; add private-state-backed `RequiresReplaceIf` for seamless migration; add all four sort settings to the adoption comparison.

## Impact

- **Users**: Existing configs using `sort_field`/`sort_order` continue to work and receive deprecation warnings. New configs can use the `sort` block for cleaner, complete sort configuration. Migrating from old to new attributes does not require destroying the index when settings are semantically equivalent, including when the plan explicitly sets Elasticsearch defaults for `missing`/`mode`. Non-equivalent `missing`/`mode` values still require destroy+recreate.
- **Code**: `internal/elasticsearch/index/index/` — schema, models, resource (Read), use_existing, and new plan modifier file (~200–250 lines across 7–8 files).
- **Docs**: `docs/resources/elasticsearch_index.md` must be regenerated after schema changes.
