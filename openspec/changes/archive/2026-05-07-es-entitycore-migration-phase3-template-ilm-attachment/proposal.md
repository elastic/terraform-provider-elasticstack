## Why

`elasticstack_elasticsearch_index_template_ilm_attachment` is already implemented on the Terraform Plugin Framework but still embeds `*entitycore.ResourceBase` directly. It duplicates the standard Read/Delete prelude and manually declares the `elasticsearch_connection` block in its schema. Migrating it to the `entitycore.ElasticsearchResource[tfModel]` envelope eliminates that duplication and brings it in line with other envelope-adopted resources.

The resource has a simple lifecycle but its Create and Update require config-derived behavior: they read the existing `@custom` component template to preserve settings/mappings/aliases, check the Elasticsearch server version, merge the ILM setting, and log warnings about version conflicts or pre-existing ILM values. These are best kept on the concrete type with placeholder envelope callbacks. Read and Delete fit the envelope callback contract cleanly.

## What Changes

- Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[tfModel]` in `internal/elasticsearch/index/templateilmattachment/resource.go`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to the `tfModel` struct.
- Extract `readILMAttachment` from the existing `Read` method body into a package-level envelope callback.
- Extract `deleteILMAttachment` from the existing `Delete` method body into a package-level envelope callback.
- Keep `Create` and `Update` on the concrete `Resource` type because they perform version-gating, content preservation, and warning logic that does not fit the simple callback contract. Pass placeholder callbacks to the envelope constructor.
- Strip the `elasticsearch_connection` block from the schema factory; the envelope injects it.
- Preserve `ImportState` passthrough on the concrete type.
- Acceptance tests depend on `index_template` and `ilm` resources from Phase 2; they remain unchanged.

## Capabilities

### New Capabilities
- `elasticsearch-index-template-ilm-attachment-via-envelope`: Migrate `elasticstack_elasticsearch_index_template_ilm_attachment` to the entitycore resource envelope while preserving version gating, content preservation, and custom ImportState.

### Modified Capabilities
<!-- None. External schema, APIs, and behavior are preserved. -->

## Impact

- Refactored code: `internal/elasticsearch/index/templateilmattachment/{resource.go,read.go,delete.go,schema.go,models.go}`.
- No public API changes. No Terraform schema changes.
- Acceptance tests for `elasticstack_elasticsearch_index_template_ilm_attachment` must continue to pass.
