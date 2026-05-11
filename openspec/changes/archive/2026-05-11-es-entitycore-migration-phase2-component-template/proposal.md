## Why

`elasticstack_elasticsearch_component_template` is the last Elasticsearch index resource still implemented on Plugin SDK v2 (`internal/elasticsearch/index/component_template.go`). Migrating it to the Terraform Plugin Framework aligns it with every other resource in the provider, enables reuse of the entitycore resource envelope, and eliminates the final SDK→PF diagnostics conversion shim in the index package.

## What Changes

- Migrate `elasticstack_elasticsearch_component_template` from Plugin SDK v2 to Plugin Framework + `entitycore.ElasticsearchResource[Data]` envelope.
- Create a new sub-package `internal/elasticsearch/index/componenttemplate` containing PF types, schema, expand/flatten logic, and envelope callbacks.
- Move resource-specific helpers out of `internal/elasticsearch/index/component_template.go` into the new package. General template helpers that remain useful for other SDK code (e.g. `template_sdk_shared.go`) stay in `internal/elasticsearch/index`.
- Implement CRUD via envelope callbacks: `readComponentTemplate`, `deleteComponentTemplate`, `createComponentTemplate`, `updateComponentTemplate`. Read uses GET, Delete uses DELETE; Create and Update both use PUT then read-back.
- Preserve the existing Terraform schema, identity format (`<cluster_uuid>/<template_name>`), import behavior, alias routing quirks, and JSON mapping semantics.
- Remove the old SDK resource file `internal/elasticsearch/index/component_template.go` and its test file once the PF resource is wired into the provider.
- Update provider registration in `provider/plugin_framework.go` (or equivalent) to register the new PF resource.

## Capabilities

### New Capabilities
- `elasticsearch-index-component-template-via-envelope`: Migrate `elasticstack_elasticsearch_component_template` to Plugin Framework and the entitycore resource envelope while preserving all externally observable behavior.

### Modified Capabilities
<!-- None. External schema, APIs, and behavior are preserved. -->

## Impact

- New code: `internal/elasticsearch/index/componenttemplate/` (resource.go, schema.go, models.go, create.go, read.go, delete.go, update.go, expand.go, flatten.go, plus tests).
- Refactored code: shared helpers in `internal/elasticsearch/index/` audited for SDK-only usage; resource-specific helpers moved.
- Deleted code: `internal/elasticsearch/index/component_template.go`, `internal/elasticsearch/index/component_template_test.go`.
- No public API changes. No Terraform schema changes. Acceptance test fixtures remain valid.
