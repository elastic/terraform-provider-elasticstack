## Why

After `terraform import`, the `elasticstack_elasticsearch_index` resource saves only `id`, `name`, `concrete_name`, `alias`, `mappings`, and `settings_raw` to state — all individual setting fields (`number_of_replicas`, `refresh_interval`, `analysis_analyzer`, slowlog thresholds, blocks, etc.) and provider-side operational attributes (`deletion_protection`, `wait_for_active_shards`, `master_timeout`, `timeout`) remain null. Every subsequent `terraform plan` shows drift for whichever of those attributes the user has configured, making import unusable in practice (issue #1587).

## What Changes

- **Override `ImportState`** to write a private state flag (`"import_hydration"`) signalling that the next read must fully hydrate all settings from the Elasticsearch API response.
- **Extend `readIndex` and `populateFromAPI`** with a `hydrateAll bool` parameter. When true (import path only), two new functions are called after the existing population logic:
  - `hydrateAllSettingsFromRaw` — parses `settings_raw` (flat JSON map with `index.*` keys) and populates every individual setting field in the model. Handles string→int64 and string→bool conversions; extracts analysis sub-categories (analyzer, tokenizer, char_filter, filter, normalizer) from the nested `index.analysis` object; handles array-valued fields (`query_default_field`) using the same pattern as the existing `extractSortSetting`. Does not touch the deprecated `settings` block or `use_existing`.
  - `populateOperationalDefaults` — sets provider-side defaults for fields that have no Elasticsearch equivalent: `deletion_protection=true`, `wait_for_active_shards="1"`, `master_timeout="30s"`, `timeout="30s"`.
- **Clear the flag** in `ModifyPlan` after pruning unconfigured Optional settings, so the behaviour is strictly scoped to the import read+plan cycle.
- **Add acceptance test coverage** verifying that a `terraform plan` immediately after `terraform import` shows no changes for an index with settings and analysis configured.

## Capabilities

### New Capabilities

_(none — this change fixes existing import behaviour, no new surface area)_

### Modified Capabilities

- `elasticsearch-index`: Adding an import-hydration requirement — after import, all individual setting fields and operational defaults SHALL be fully populated in state so that a subsequent plan produces no changes when the config matches the imported index.

## Impact

- **Code**: `internal/elasticsearch/index/index/` — `constants.go`, `resource.go`, `read.go`, `models.go`, `create.go`, plus a new `settings_read.go`.
- **Private state**: A new key `"import_hydration"` is written to provider private state during `ImportState` and cleared in the following `ModifyPlan`. It does not affect the `sort_config` key or any other existing private state.
- **Behaviour**: Regular read/plan/apply cycles are unaffected — the `hydrateAll` flag is only set on the import path.
- **No breaking changes**.
