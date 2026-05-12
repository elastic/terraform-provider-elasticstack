## Why

The `elasticstack_elasticsearch_security_user` data source is one of the last remaining SDK-based Elasticsearch data sources. Migrating it to Plugin Framework and wrapping it with `entitycore.NewElasticsearchDataSource` removes the Terraform Plugin SDK v2 dependency for this entity, aligns it with the provider-wide PF envelope pattern, and eliminates duplicated connection-resolution and state-persistence boilerplate.

## What Changes

- Rewrite `internal/elasticsearch/security/user_data_source.go` from a `*schema.Resource` SDK implementation to a Plugin Framework `datasource.DataSource` using `entitycore.NewElasticsearchDataSource`.
- Introduce a PF model struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields matching the current schema (`username`, `full_name`, `email`, `roles`, `metadata`, `enabled`).
- Convert the SDK schema to a `schema.Schema` with Plugin Framework attribute types (`schema.StringAttribute`, `schema.BoolAttribute`, `schema.SetAttribute` with `ElementType: types.StringType`).
- Move the API call and state-mapping logic into a `readDataSource` callback returning `(model, diag.Diagnostics)`.
- Register the new PF constructor in `provider/plugin_framework.go` and remove the SDK registration from `provider/provider.go`.
- Update or replace acceptance tests to use PF testing patterns.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-security-user`: The data source implementation SHALL migrate from Terraform Plugin SDK v2 to Plugin Framework and SHALL use `entitycore.NewElasticsearchDataSource` for connection handling, config decode, and state persistence. The resource portion of this capability is unaffected.

## Impact

- `internal/elasticsearch/security/user_data_source.go` — complete rewrite to PF
- `provider/provider.go` — remove SDK data source registration
- `provider/plugin_framework.go` — add PF data source registration
- `internal/elasticsearch/security/user_data_source_test.go` — migrate tests to PF patterns if needed
