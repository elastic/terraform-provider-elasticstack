## MODIFIED Requirements

### Requirement: Kibana resource envelope injects the resource-schema connection block

`entitycore.NewKibanaResource` SHALL inject `internal/schema.GetKbResourceConnectionBlock()` into the resource schema instead of `internal/schema.GetKbFWConnectionBlock()`.

#### Scenario: Managed Kibana resource schema uses the resource-schema block

- **WHEN** `NewKibanaResource` constructs a resource
- **THEN** its schema SHALL include the `kibana_connection` block from `GetKbResourceConnectionBlock()`

### Requirement: Kibana resource envelope resolves resource-variant connections

`entitycore.KibanaResource` SHALL continue to invoke `clients.ProviderClientFactory.GetKibanaClient` to resolve scoped Kibana clients. The `GetKibanaClient` method SHALL decode the incoming `types.List` into `[]internal/clients/config.KibanaResourceConnection`, apply `_wo`-over-plain preference to produce `[]internal/clients/config.KibanaConnection`, and build the client from the resolved connection. This applies to all CRUD/ImportState paths, including both the `getClient` closure and the direct `runKibanaWrite` call.

#### Scenario: Create and Update use `_wo` credentials

- **WHEN** a resource is created or updated with `api_key_wo` in `kibana_connection`
- **THEN** `GetKibanaClient` SHALL authenticate using the `api_key_wo` value

### Requirement: Kibana resource envelope implements ModifyPlan for `_wo` drift detection

`entitycore.KibanaResource` SHALL satisfy `resource.ResourceWithModifyPlan`. The envelope SHALL use `internal/utils/writeonlyhash` to detect silent in-config changes to each `_wo` credential attribute in the `kibana_connection` block (`password_wo`, `api_key_wo`, `bearer_token_wo`), with one `Hasher` per concrete resource type. The envelope `ModifyPlan` SHALL be a no-op when no `_wo` attribute is configured.

Concrete resources that already define their own `ModifyPlan` method SHALL delegate to the envelope's `ModifyPlan` or otherwise ensure `_wo` drift detection runs.

#### Scenario: Envelope ModifyPlan schedules update on changed `_wo` value

- **WHEN** the configured `api_key_wo` value differs from the value last applied
- **THEN** `ModifyPlan` SHALL emit a warning naming the attribute path
- **AND** an update SHALL be scheduled

#### Scenario: Envelope ModifyPlan is a no-op without `_wo` attributes

- **WHEN** `kibana_connection` is not configured or no `_wo` attribute is set
- **THEN** `ModifyPlan` SHALL NOT read, write, or clear any private-state keys
