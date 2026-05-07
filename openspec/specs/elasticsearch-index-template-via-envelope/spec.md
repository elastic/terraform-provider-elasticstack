# elasticsearch-index-template-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase2-template-ilm. Update Purpose after archive.
## Requirements
### Requirement: Index template resource uses the entitycore envelope
The `elasticstack_elasticsearch_index_template` resource SHALL embed `*entitycore.ElasticsearchResource[Model]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, `resource.ResourceWithImportState`, `resource.ResourceWithModifyPlan`, `resource.ResourceWithValidateConfig`, and `resource.ResourceWithUpgradeState`.

#### Scenario: Resource is registered as an envelope resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_index_template` SHALL be constructed via `entitycore.NewElasticsearchResource[Model]`
- **AND** the concrete type SHALL declare `UpgradeState`, `ModifyPlan`, `ValidateConfig`, and `ImportState`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all attributes and blocks SHALL match the pre-migration schema exactly

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `Model` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model getters return correct fields
- **WHEN** a `Model` value is created with `ID`, `Name`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `Name` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Read callback performs alias reconciliation and canonicalization
The envelope's read callback SHALL fetch the template from the API, apply alias reconciliation using the prior-state model, canonicalize the alias set, copy `ID` and `ElasticsearchConnection` from the prior state, and return the prepared model.

#### Scenario: Read reconciles alias routing from prior state
- **GIVEN** a prior state with alias `routing = "x"`
- **WHEN** the API response omits routing fields
- **THEN** the returned model SHALL retain the prior alias configuration via reconciliation

#### Scenario: Read canonicalizes alias set elements
- **GIVEN** an API response with aliases containing equivalent routing values
- **WHEN** the read callback runs
- **THEN** the returned model SHALL have a canonical alias set

### Requirement: Create and Update remain on the concrete type
The concrete `Resource` type SHALL override `Create` and `Update` to preserve config-derived behavior: alias reconciliation from configuration, server-version gating, and the 8.x `allow_custom_routing` workaround.

#### Scenario: Create uses config for alias reconciliation
- **GIVEN** a create plan with aliases configured
- **WHEN** create runs
- **THEN** it SHALL read `req.Config` to reconcile alias values after the read-back

#### Scenario: Update applies allow_custom_routing workaround
- **GIVEN** prior state had `allow_custom_routing = true` and config removes the attribute
- **WHEN** update runs
- **THEN** the PUT body SHALL include `allow_custom_routing = false`

### Requirement: State upgrader is preserved
The concrete type SHALL continue to register the V0→V1 state upgrader, collapsing list-shaped blocks to `SingleNestedBlock` object shapes.

#### Scenario: Upgrade from SDK-shaped state
- **GIVEN** state written by the Plugin SDK v2 implementation
- **WHEN** Terraform refreshes with the new implementation
- **THEN** the V0→V1 upgrader SHALL run
- **AND** no diff SHALL be produced

### Requirement: ModifyPlan and ValidateConfig are preserved
The concrete type SHALL continue to implement `ModifyPlan` and `ValidateConfig` with the same logic as before migration.

#### Scenario: ModifyPlan suppresses semantic drift for settings
- **GIVEN** planned settings differ from state only by JSON key form
- **WHEN** `ModifyPlan` runs
- **THEN** the plan SHALL be adjusted so no diff is reported

#### Scenario: ValidateConfig rejects data_stream_options without failure_store
- **GIVEN** `template.data_stream_options` is configured without `failure_store`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a plan-time error

