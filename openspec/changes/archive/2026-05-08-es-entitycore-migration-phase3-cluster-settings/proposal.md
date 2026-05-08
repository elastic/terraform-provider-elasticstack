## Why

`elasticstack_elasticsearch_cluster_settings` is one of the last remaining Plugin SDK resources in the Elasticsearch cluster package. It duplicates provider-wide patterns (connection-block handling, composite-ID construction, state persistence) that the entitycore resource envelope already centralizes for Plugin Framework resources. Rewriting it to Plugin Framework and adopting the envelope removes SDK tech debt, unifies error handling and schema construction with modern resources, and collapses duplicated boilerplate.

## What Changes

- Rewrite `internal/elasticsearch/cluster/settings.go` from Plugin SDK to Plugin Framework.
- Introduce a new `internal/elasticsearch/cluster/settings/` package (or keep the file in place) with PF types: a `tfModel` struct, schema factory, model conversion helpers, and callback functions.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to the new model so it satisfies `ElasticsearchResourceModel`.
- Use `PlaceholderElasticsearchWriteCallbacks` for create and update because those write paths need special handling:
  - Create and Update both PUT settings, but Update additionally needs the prior state to null out removed settings.
  - Delete needs the current state to know which setting keys to remove, so the concrete Delete override delegates to the same helper passed as the envelope's required delete callback.
- Override `Create`, `Update`, and `Delete` on the concrete resource type. Let the envelope own `Read`, `Schema`, `Configure`, and `Metadata`.
- Preserve the existing `persistent` / `transient` block schema, setting name/value/value_list semantics, list-vs-scalar type detection, and import support.

No external Terraform schema or acceptance-test changes.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-cluster-settings`: Implementation rewritten from Plugin SDK to Plugin Framework + entitycore envelope. Externally-observable behavior preserved verbatim.

## Impact

- Rewritten code: `internal/elasticsearch/cluster/settings.go` becomes PF-based.
- New PF model, schema, and conversion code in the same package.
- The resource preserves its existing Terraform type name `elasticstack_elasticsearch_cluster_settings`.
- Acceptance tests must continue to pass without modification.
