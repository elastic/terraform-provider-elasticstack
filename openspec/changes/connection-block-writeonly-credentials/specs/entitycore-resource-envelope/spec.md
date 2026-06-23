## MODIFIED Requirements

### Requirement: Elasticsearch resource envelope injects the resource-schema connection block

`entitycore.NewElasticsearchResource` SHALL inject `internal/schema.GetEsResourceConnectionBlock()` into the resource schema instead of `internal/schema.GetEsFWConnectionBlock()`.

#### Scenario: Managed Elasticsearch resource schema uses the resource-schema block

- **WHEN** `NewElasticsearchResource` constructs a resource
- **THEN** its schema SHALL include the `elasticsearch_connection` block from `GetEsResourceConnectionBlock()`

### Requirement: Elasticsearch resource envelope resolves resource-variant connections

`entitycore.ElasticsearchResource` SHALL continue to invoke `clients.ProviderClientFactory.GetElasticsearchClient` to resolve scoped Elasticsearch clients. The `GetElasticsearchClient` method SHALL decode the incoming `types.List` into `[]internal/clients/config.ElasticsearchResourceConnection`, apply `_wo`-over-plain preference to produce `[]internal/clients/config.ElasticsearchConnection`, and build the client from the resolved connection. This applies to all CRUD/ImportState paths, including both the `getClient` closure and the direct `runWrite` call.

#### Scenario: Create and Update use `_wo` credentials

- **WHEN** a resource is created or updated with `password_wo` in `elasticsearch_connection`
- **THEN** `GetElasticsearchClient` SHALL authenticate using the `password_wo` value

### Requirement: Elasticsearch resource envelope implements ModifyPlan for `_wo` drift detection

`entitycore.ElasticsearchResource` SHALL satisfy `resource.ResourceWithModifyPlan`. The envelope SHALL use `internal/utils/writeonlyhash` to detect silent in-config changes to each `_wo` credential attribute in the `elasticsearch_connection` block, with one `Hasher` per concrete resource type. The envelope `ModifyPlan` SHALL be a no-op when no `_wo` attribute is configured.

Concrete resources that already define their own `ModifyPlan` method SHALL delegate to the envelope's `ModifyPlan` or otherwise ensure `_wo` drift detection runs.

#### Scenario: Envelope ModifyPlan schedules update on changed `_wo` value

- **WHEN** the configured `password_wo` value differs from the value last applied
- **THEN** `ModifyPlan` SHALL emit a warning naming the attribute path
- **AND** an update SHALL be scheduled

#### Scenario: Envelope ModifyPlan is a no-op without `_wo` attributes

- **WHEN** `elasticsearch_connection` is not configured or no `_wo` attribute is set
- **THEN** `ModifyPlan` SHALL NOT read, write, or clear any private-state keys
