## MODIFIED Requirements

### Requirement: Resource-level `kibana_connection` block uses the resource-schema variant

The managed-resource `kibana_connection` block surface SHALL be extended by the `kibana-connection-writeonly` capability and SHALL be provided through `internal/schema.GetKbResourceConnectionBlock()`. The block SHALL retain every attribute present in `internal/schema.GetKbFWConnectionBlock()` and MAY add the write-only companions defined by `kibana-connection-writeonly`.

Provider-level, data-source, ephemeral-resource, and action-connection surfaces SHALL continue to use `internal/schema.GetKbFWConnectionBlock()` / `internal/schema.GetKbActionConnectionBlock()` unchanged.

#### Scenario: Managed resource exposes the resource-schema block

- **WHEN** a managed resource is constructed via `entitycore.NewKibanaResource`
- **THEN** its schema SHALL include a `kibana_connection` block produced by `GetKbResourceConnectionBlock()`
- **AND** the block SHALL include the `_wo` credential companions defined by `kibana-connection-writeonly`

#### Scenario: Provider-level block remains unchanged

- **WHEN** the provider's `kibana` block is configured
- **THEN** it SHALL continue to use `GetKbFWConnectionBlock()` and SHALL NOT expose `_wo` attributes

#### Scenario: Data source and ephemeral blocks remain unchanged

- **WHEN** a data source or ephemeral resource exposes a `kibana_connection` block
- **THEN** the block SHALL be provided by `GetKbFWConnectionBlock()` or `GetKbEphemeralConnectionBlock()` and SHALL NOT expose `_wo` attributes
