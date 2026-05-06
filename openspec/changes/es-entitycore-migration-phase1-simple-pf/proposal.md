## Why

Migrate four simple Plugin Framework resources that still embed `*entitycore.ResourceBase` directly to `*entitycore.ElasticsearchResource[Data]`. Each has straightforward CRUD lifecycles that fit the envelope callback contract completely, eliminating duplicated Read/Delete/Schema/client-resolution code.

## What Changes

Migrate to `NewElasticsearchResource` envelope with full callback wiring:
- `elasticstack_elasticsearch_index_alias` (`internal/elasticsearch/index/alias`)
- `elasticstack_elasticsearch_data_stream_lifecycle` (`internal/elasticsearch/index/datastreamlifecycle`)
- `elasticstack_elasticsearch_enrich_policy` (`internal/elasticsearch/enrich`)
- `elasticstack_elasticsearch_inference_endpoint` (`internal/elasticsearch/inference/inferenceendpoint`)

For each:
- Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` on the model.
- Extract `Read`/`Delete` bodies into package-level `read*` and `delete*` callbacks.
- Extract `Create`/`Update` into `create*` and `update*` callbacks (where they currently fit the envelope contract).
- Strip manual `elasticsearch_connection` block from schema factories; envelope injects it.
- Preserve `ImportState` passthrough or custom import logic on concrete types.
- Preserve any `UpgradeState` by keeping the method on the concrete type.

## Capabilities

### New Capabilities
- `index-alias-resource-via-envelope`
- `data-stream-lifecycle-resource-via-envelope`
- `enrich-policy-resource-via-envelope`
- `inference-endpoint-resource-via-envelope`

### Modified Capabilities
<!-- None. -->

## Impact

- Four packages affected; no public Terraform interface changes.
- Reduces ~300+ lines of duplicated CRUD boilerplate across these packages.
