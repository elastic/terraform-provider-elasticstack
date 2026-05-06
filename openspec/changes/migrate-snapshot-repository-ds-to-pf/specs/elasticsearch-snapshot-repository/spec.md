## ADDED Requirements

### Requirement: Data source uses Plugin Framework and entitycore envelope

The data source SHALL be implemented as a Plugin Framework `datasource.DataSource` constructed via `entitycore.NewElasticsearchDataSource`. The concrete model SHALL embed `entitycore.ElasticsearchConnectionField` and SHALL satisfy `entitycore.ElasticsearchDataSourceModel`. The envelope SHALL own config decode, scoped client resolution, and state persistence.

#### Scenario: Envelope handles connection and decode

- **WHEN** the data source is evaluated
- **THEN** `entitycore.NewElasticsearchDataSource` SHALL decode the configuration into the concrete model
- **AND** resolve the scoped Elasticsearch client from the model's `elasticsearch_connection` block
- **AND** invoke the entity-specific read callback
- **AND** persist the returned model to state

#### Scenario: Read callback owns API call and id assignment

- **WHEN** the entity-specific read callback is invoked with the scoped client and config
- **THEN** it SHALL call `elasticsearch.GetSnapshotRepository`
- **AND** set `model.ID` to `<cluster_uuid>/<repository_name>`
- **AND** map the API response into the corresponding type block in the model

## MODIFIED Requirements

### Requirement: Data source read-only semantics (REQ-DS-001)

The data source SHALL support only a read operation. It SHALL NOT perform create, update, or delete operations. The data source SHALL be constructed via `entitycore.NewElasticsearchDataSource`.

#### Scenario: Read-only data source

- **GIVEN** the data source is configured
- **WHEN** Terraform evaluates the data source
- **THEN** the provider SHALL only read the repository and SHALL NOT create, update, or delete it

### Requirement: Data source API (REQ-DS-002)

The data source SHALL use the Elasticsearch Get Snapshot Repository API (`GET /_snapshot/<repository>`) to fetch the repository identified by `name`. When the API returns a non-success status, the data source SHALL surface the API error to Terraform diagnostics. When the repository is not found (API returns `nil` with no error), the data source SHALL set `id`, return a warning diagnostic with the message "Could not find snapshot repository [<name>]", and SHALL not attempt to populate type block attributes.

#### Scenario: Repository not found

- **GIVEN** no repository with the requested name exists
- **WHEN** the data source is read
- **THEN** a warning diagnostic SHALL be returned and type block attributes SHALL remain empty

### Requirement: Data source identity (REQ-DS-003)

The data source SHALL set `id` in the format `<cluster_uuid>/<repository_name>` by calling `client.ID(ctx, repoName)` after resolving the client. The `id` SHALL be set regardless of whether the repository was found.

#### Scenario: Data source id set

- **GIVEN** the data source read runs for a repository name
- **WHEN** the provider resolves the Elasticsearch client
- **THEN** `id` SHALL be set to `<cluster_uuid>/<repository_name>`

### Requirement: Data source connection (REQ-DS-004)

The data source SHALL resolve a `*clients.ElasticsearchScopedClient` from the provider client factory. When `elasticsearch_connection` is absent, the factory SHALL return a typed client built from provider-level defaults. When `elasticsearch_connection` is configured, the factory SHALL return a typed scoped client rebuilt from that connection. Connection resolution SHALL be owned by the `entitycore.NewElasticsearchDataSource` envelope.

#### Scenario: Data source-scoped connection

- **GIVEN** `elasticsearch_connection` is configured on the data source
- **WHEN** the data source reads the repository
- **THEN** the provider SHALL use the typed scoped client rebuilt from that connection

### Requirement: Data source type block population (REQ-DS-005)

After a successful read, the data source SHALL set the `type` attribute to the repository type string returned by the API. The data source SHALL populate only the type block corresponding to the returned type; all other type blocks SHALL remain empty. The data source SHALL flatten settings from the API response using the same type conversion logic as the resource (string-to-int, string-to-bool, string-as-string). If the `type` returned by the API does not match any of the supported type block names in the schema, the data source SHALL return an error diagnostic.

#### Scenario: GCS repository

- **GIVEN** a GCS snapshot repository exists in Elasticsearch
- **WHEN** the data source is read
- **THEN** `type` SHALL be `"gcs"`, the `gcs` block SHALL be populated with the repository settings, and all other type blocks SHALL remain empty

### Requirement: Data source schema — computed attributes (REQ-DS-006)

All attributes in the data source schema except `name` SHALL be computed. The `name` attribute SHALL be required. The data source schema does NOT include `max_number_of_snapshots` for the `gcs`, `azure`, `s3`, and `hdfs` type blocks; only `fs` and `url` merge `commonStdSettings` and therefore include that attribute. The S3 type block in the data source SHALL NOT include the `endpoint` attribute.

#### Scenario: Name is required

- **GIVEN** no `name` is provided in the data source configuration
- **WHEN** Terraform validates the configuration
- **THEN** a validation error SHALL be returned
