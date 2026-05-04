## REMOVED Requirements

### Requirement: Existing struct-based data sources remain functional
**Reason**: The `DataSourceBase` struct and struct-based embedding pattern have been removed. All Plugin Framework data sources now use the generic envelope constructors. Removing `DataSourceBase` eliminates the dual-pattern maintenance burden and unifies the codebase on the envelope.
**Migration**: Any new or in-flight Plugin Framework data sources should use `entitycore.NewKibanaDataSource[T]` or `entitycore.NewElasticsearchDataSource[T]` instead of embedding `*entitycore.DataSourceBase`.

## ADDED Requirements

### Requirement: Envelope is the sole supported Plugin Framework data source pattern
The system SHALL provide only the generic envelope constructors for Plugin Framework data source wiring. No struct-based alternative SHALL exist within `internal/entitycore`. All existing Plugin Framework data sources SHALL be constructed via `NewKibanaDataSource` or `NewElasticsearchDataSource`.

#### Scenario: All PF data sources use envelope constructor
- **WHEN** the provider enumerates its Plugin Framework data sources
- **THEN** every data source SHALL be constructed via `entitycore.NewKibanaDataSource[T]` or `entitycore.NewElasticsearchDataSource[T]`
- **AND** no data source SHALL embed `*entitycore.DataSourceBase`

#### Scenario: DataSourceBase is not present in the entitycore package
- **WHEN** developers inspect `internal/entitycore` for data source base types
- **THEN** `DataSourceBase` SHALL NOT exist
- **AND** only envelope constructors and connection field embeddable structs SHALL be available

### Requirement: Kibana data source models embed KibanaConnectionField
Kibana-backed envelope data source models SHALL embed `entitycore.KibanaConnectionField` (or an equivalent struct field with the `GetKibanaConnection` accessor) so the envelope can decode the `kibana_connection` block alongside entity attributes.

#### Scenario: Model embedding satisfies KibanaDataSourceModel
- **WHEN** a data source model embeds `KibanaConnectionField`
- **THEN** the model SHALL satisfy the `KibanaDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `kibana_connection` block into that field during `Read`

### Requirement: Elasticsearch data source models embed ElasticsearchConnectionField
Elasticsearch-backed envelope data source models SHALL embed `entitycore.ElasticsearchConnectionField` (or an equivalent struct field with the `GetElasticsearchConnection` accessor) so the envelope can decode the `elasticsearch_connection` block alongside entity attributes.

#### Scenario: Model embedding satisfies ElasticsearchDataSourceModel
- **WHEN** a data source model embeds `ElasticsearchConnectionField`
- **THEN** the model SHALL satisfy the `ElasticsearchDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `elasticsearch_connection` block into that field during `Read`
