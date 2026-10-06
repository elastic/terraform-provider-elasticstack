## Why

Users who manage concrete Elasticsearch indices through `elasticstack_elasticsearch_index_template` and ILM rollover (rather than through `elasticstack_elasticsearch_index` directly) have no way to update dynamic index settings on an *existing* concrete index. When an index template's settings change (for example raising `index.mapping.total_fields.limit`), already-created backing/concrete indices do not inherit the change, and subsequent `elasticstack_elasticsearch_index_mappings` applies can fail with errors such as:

```text
Error: status: 400, failed: [illegal_argument_exception], reason: Limit of total fields [2300] has been exceeded
```

Adopting the index into `elasticstack_elasticsearch_index` is not viable: it would make Terraform own index lifecycle concerns (shards, aliases, mappings) that are already managed by the template/ILM pipeline, risking conflicts.

`elasticstack_elasticsearch_index_mappings` already solves the equivalent problem for mappings: manage a declared subset on an existing index without adopting the whole index. This change adds the settings analog, `elasticstack_elasticsearch_index_settings`, scoped to a single concrete index per resource instance.

## What Changes

- Add a new resource, `elasticstack_elasticsearch_index_settings`, that manages a user-declared subset of **dynamic** (runtime-changeable) index settings on an existing concrete Elasticsearch index, without taking ownership of the index's lifecycle.
- Extract the dynamic-setting attribute definitions currently hand-declared in `internal/elasticsearch/index/index/schema.go` into a single shared function, `GetDynamicSettingAttributes() map[string]schema.Attribute`, owned by `internal/elasticsearch/index/settings_keys.go`. Both `elasticstack_elasticsearch_index` and the new `elasticstack_elasticsearch_index_settings` resource consume this shared definition — merged into each resource's own schema map — instead of hand-duplicating ~30 attribute declarations. This migration lands in the same PR as the new resource.
- Introduce a shared, embeddable `dynamicsettings.Model` struct (one field per `DynamicSettingsKeys` entry) that both resources' `tfModel`s embed anonymously, and extend `internal/elasticsearch/index/reflectutil.go`'s `GetFieldValueByTagValue`/`SetFieldValueByTagValue` to recurse into anonymous (embedded) struct fields so the existing `index` resource's reflection-based settings round-trip (`toIndexSettings`, `hydrateAllSettingsFromRaw`) keeps working unchanged against the embedded fields.
- Add a `settings_json` escape-hatch attribute (`jsontypes.Normalized`) on the new resource for dynamic settings not yet represented as a typed attribute. It is permissive on unknown keys (so future Elasticsearch dynamic settings work without a provider release) but rejects, at plan time, any key that is a literal `StaticSettingsKeys` match (settings that can only be set at index-creation time).
- Scope v1 to exactly one concrete, already-resolved index name per resource instance (no wildcard-to-many-indices state modeling); multiple concrete indices are handled by the caller's own `for_each`, per the issue's example.
- `id` follows the existing `<cluster_uuid>/<index_name>` convention used by `elasticstack_elasticsearch_index_mappings`.
- Destroy is a fixed no-op: removing the resource from configuration never reverts or clears settings on the live index. (No configurable destroy behavior in v1.)
- Read reconciles only the subset of settings declared in prior state (plus the escape hatch), mirroring the mapping resource's "declared subset" read model — unrelated settings returned by Elasticsearch are ignored and never cause drift.
- Setting a previously-declared attribute to `null` on a subsequent apply resets that setting via Elasticsearch's null-to-reset API semantics, consistent with `elasticstack_elasticsearch_index`'s existing `update.go` behavior (`updateSettings`), rather than merely dropping it from tracking.

## Capabilities

### New Capabilities

- `elasticstack-elasticsearch-index-settings`: the new `elasticstack_elasticsearch_index_settings` resource — schema, create/read/update, no-op delete, identity/import, and the `settings_json` escape hatch with static-key rejection.

### Modified Capabilities

- `elasticsearch-index`: `elasticstack_elasticsearch_index`'s dynamic-setting attributes are now sourced from the shared `GetDynamicSettingAttributes()` function and the embedded `dynamicsettings.Model` rather than hand-declared per-attribute. Attribute names, types, descriptions, and existing plan-modifier/validation behavior are preserved — this is an internal refactor, not a behavior change, and is covered by requirement REQ-UNCHANGED in the capability's delta spec.

## Impact

- `internal/elasticsearch/index/settings_keys.go` — gains `GetDynamicSettingAttributes() map[string]schema.Attribute`, the single owned definition of type/description/plan-modifiers per `DynamicSettingsKeys` entry.
- `internal/elasticsearch/index/dynamicsettings/` (new package) — shared `Model` struct over `DynamicSettingsKeys`, embedded anonymously by both resources' `tfModel`s.
- `internal/elasticsearch/index/reflectutil.go` — `GetFieldValueByTagValue`/`SetFieldValueByTagValue` extended to recurse into anonymous embedded struct fields.
- `internal/elasticsearch/index/index/schema.go`, `models.go` — migrate dynamic-setting attributes to `GetDynamicSettingAttributes()` and embed `dynamicsettings.Model` in `tfModel`; existing attribute behavior must not change for current users.
- `internal/elasticsearch/index/indexsettings/` (new package) — new resource following the `indexmappings` package's envelope pattern (`entitycore.NewElasticsearchResource`, `WriteRequest`/`WriteResult`, no-op `Delete`, `ImportStatePassthroughID`).
- `internal/clients/elasticsearch/index.go` — reuses existing `GetIndex`/`UpdateIndexSettings`; no new client code expected.
- `internal/elasticsearch/queryrulesets/validators.go` pattern reused for a new `settings_json` plan-time validator rejecting literal `StaticSettingsKeys` matches.
- Provider registration, docs, and acceptance tests for the new resource.
