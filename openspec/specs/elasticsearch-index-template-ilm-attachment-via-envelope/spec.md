# elasticsearch-index-template-ilm-attachment-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase3-template-ilm-attachment. Update Purpose after archive.
## Requirements
### Requirement: Index template ILM attachment resource uses the entitycore envelope
The `elasticstack_elasticsearch_index_template_ilm_attachment` resource SHALL embed `*entitycore.ElasticsearchResource[tfModel]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.

#### Scenario: Resource is registered as an envelope resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_index_template_ilm_attachment` SHALL be constructed via `entitycore.NewElasticsearchResource[tfModel]`
- **AND** the concrete type SHALL declare `ImportState`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all attributes SHALL match the pre-migration schema exactly

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `tfModel` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model getters return correct fields
- **WHEN** a `tfModel` value is created with `ID`, `IndexTemplate`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the derived component template name (`IndexTemplate + "@custom"`)
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Read callback derives index_template on import
The envelope's read callback SHALL parse the composite ID, derive the component template name, and when `index_template` is unknown derive it by stripping the `@custom` suffix. It SHALL call Get Component Template with `flat_settings=true`. It SHALL return `found == false` when the template or the ILM setting is missing.

#### Scenario: Import derives index_template from id
- **GIVEN** an import with id `<cluster_uuid>/logs-system.syslog@custom`
- **WHEN** read runs after import
- **THEN** `index_template` SHALL be set to `"logs-system.syslog"`

#### Scenario: ILM setting absent removes from state
- **GIVEN** a `@custom` component template that exists but has no `index.lifecycle.name` setting
- **WHEN** read runs
- **THEN** the resource SHALL be removed from state

### Requirement: Create and Update remain on the concrete type
The concrete `Resource` type SHALL override `Create` and `Update` to preserve version-gating, content preservation, and warning behavior.

#### Scenario: Create rejects old Elasticsearch versions
- **GIVEN** Elasticsearch version < 8.2.0
- **WHEN** create runs
- **THEN** the provider SHALL return an "Unsupported Elasticsearch Version" error and not call the API

#### Scenario: Update preserves existing template content
- **GIVEN** an existing `@custom` component template with custom mappings
- **WHEN** update runs to change the ILM policy
- **THEN** the existing mappings SHALL be preserved and only `index.lifecycle.name` SHALL be updated

### Requirement: Delete callback removes ILM setting via Put
The envelope's delete callback SHALL read the existing `@custom` component template, remove the `index.lifecycle.name` key from settings, and write the updated template via Put Component Template. If the template does not exist, it SHALL return nil diagnostics without calling the API.

#### Scenario: Delete updates rather than deletes
- **GIVEN** an existing `@custom` component template
- **WHEN** delete runs
- **THEN** the resource SHALL call Put Component Template with the ILM setting removed
- **AND** it SHALL NOT call Delete Component Template

#### Scenario: Empty settings pruned on delete
- **GIVEN** the only setting in the template is `index.lifecycle.name`
- **WHEN** delete runs and removes the ILM setting
- **THEN** the resource SHALL write the template with a nil settings map

### Requirement: Import behavior is preserved
The concrete resource type SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import with composite id
- **GIVEN** an import identifier in the format `<cluster_uuid>/<index_template>@custom`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state for subsequent operations

