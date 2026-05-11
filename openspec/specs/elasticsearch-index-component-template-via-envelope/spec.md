# `elasticstack_elasticsearch_component_template` — Entitycore Envelope Requirements

## Purpose

Define the entitycore resource envelope behavior for the `elasticstack_elasticsearch_component_template` resource: envelope-based implementation, model interfaces, callback contracts, alias routing preservation, JSON field validation, and import behavior.

## Schema

See `openspec/specs/elasticsearch-index-component-template/spec.md` for the full schema.

## Requirements

### Requirement: Component template resource uses the entitycore envelope

The `elasticstack_elasticsearch_component_template` resource SHALL be implemented on the Terraform Plugin Framework and SHALL embed `*entitycore.ElasticsearchResource[Data]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.

#### Scenario: Resource is registered as a Plugin Framework resource

- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_component_template` SHALL be registered as a Plugin Framework resource
- **AND** it SHALL be constructed via `entitycore.NewElasticsearchResource[Data]`

#### Scenario: Schema includes injected connection block

- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all other attributes and blocks SHALL match the legacy SDK schema shape

### Requirement: Model satisfies ElasticsearchResourceModel

The resource's `Data` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` as value-receiver methods.

#### Scenario: Model getters return correct fields

- **WHEN** a `Data` value is created with `ID`, `Name`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `Name` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Read callback fetches and maps component templates

The envelope's `readFunc` SHALL call the Get Component Template API, return `(model, true, nil)` when the template exists, and return `(_, false, nil)` when the template is not found (404).

#### Scenario: Successful read returns populated model

- **GIVEN** a component template exists in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated `Data` with `name`, `version`, `metadata`, and `template` fields set from the API response

#### Scenario: Missing template removes from state

- **GIVEN** the component template does not exist in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`
- **AND** the envelope SHALL remove the resource from state

### Requirement: Delete callback removes component templates

The envelope's `deleteFunc` SHALL call the Delete Component Template API with the parsed resource identifier.

#### Scenario: Successful delete

- **GIVEN** a component template exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call Delete Component Template API
- **AND** return nil diagnostics on success

### Requirement: Create and update callbacks use Put Component Template API

The `createFunc` and `updateFunc` callbacks SHALL construct a `models.ComponentTemplate` from the planned model, call the Put Component Template API, compute the composite `id`, and return the written model. The envelope SHALL invoke `readFunc` after a successful callback and persist the read result to state.

#### Scenario: Create puts template and refreshes state

- **GIVEN** a valid planned `Data` model
- **WHEN** the create callback runs
- **THEN** it SHALL call Put Component Template API
- **AND** the envelope SHALL read back the template and set state from the read result

#### Scenario: Update puts template and refreshes state

- **GIVEN** an existing component template and a changed planned model
- **WHEN** the update callback runs
- **THEN** it SHALL call Put Component Template API
- **AND** the envelope SHALL read back the template and set state from the read result

### Requirement: Alias routing is preserved on refresh

On read, user-defined alias `routing` SHALL be preserved when the API omits it, matching the legacy SDK behavior.

#### Scenario: Routing preserved after read-back

- **GIVEN** a configuration with alias `routing = "x"`
- **WHEN** the API response omits the routing field
- **THEN** the refreshed state SHALL retain the user-configured `routing` value

### Requirement: JSON fields are validated and normalized

`metadata`, `template.mappings`, `template.settings`, and `template.alias.filter` SHALL be validated as JSON during plan validation and normalized during create/update/read.

#### Scenario: Invalid JSON rejected at plan time

- **GIVEN** `template.mappings` contains invalid JSON
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation error

### Requirement: Import behavior is preserved

The concrete resource type SHALL implement `ImportState` as a passthrough on the `id` attribute, preserving the legacy import format `<cluster_uuid>/<template_name>`.

#### Scenario: Import with composite id

- **GIVEN** an import identifier in the format `<cluster_uuid>/<template_name>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state for subsequent read and delete operations