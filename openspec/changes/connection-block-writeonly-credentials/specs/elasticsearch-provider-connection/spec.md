## MODIFIED Requirements

### Requirement: Resource-level `elasticsearch_connection` block uses the resource-schema variant

The managed-resource `elasticsearch_connection` block surface SHALL be extended by the `elasticsearch-connection-writeonly` capability and SHALL be provided through `internal/schema.GetEsResourceConnectionBlock()`. The block SHALL retain every attribute present in `internal/schema.GetEsFWConnectionBlock()` and MAY add the write-only companions defined by `elasticsearch-connection-writeonly`.

Provider-level, data-source, ephemeral-resource, and action-connection surfaces SHALL continue to use `internal/schema.GetEsFWConnectionBlock()` / `internal/schema.GetEsActionConnectionBlock()` unchanged.

#### Scenario: Managed resource exposes the resource-schema block

- **WHEN** a managed resource is constructed via `entitycore.NewElasticsearchResource`
- **THEN** its schema SHALL include an `elasticsearch_connection` block produced by `GetEsResourceConnectionBlock()`
- **AND** the block SHALL include the `_wo` credential companions defined by `elasticsearch-connection-writeonly`

#### Scenario: Provider-level block remains unchanged

- **WHEN** the provider's `elasticsearch` block is configured
- **THEN** it SHALL continue to use `GetEsFWConnectionBlock()` and SHALL NOT expose `_wo` attributes

#### Scenario: Data source and ephemeral blocks remain unchanged

- **WHEN** a data source or ephemeral resource exposes an `elasticsearch_connection` block
- **THEN** the block SHALL be provided by `GetEsFWConnectionBlock()` or `GetEsEphemeralConnectionBlock()` and SHALL NOT expose `_wo` attributes
