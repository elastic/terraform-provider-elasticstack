## ADDED Requirements

### Requirement: Cluster settings resource uses the entitycore envelope for Schema and Read
The `elasticstack_elasticsearch_cluster_settings` resource SHALL embed `*entitycore.ElasticsearchResource[tfModel]`. The envelope SHALL own Schema and Read.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create override expands and puts settings
The concrete resource SHALL override `Create` to expand persistent and transient settings and PUT them.

#### Scenario: Create cluster settings
- **GIVEN** a planned model with persistent settings
- **WHEN** create runs
- **THEN** it SHALL expand settings and PUT to Elasticsearch
- **AND** flatten and set state from the response

### Requirement: Update override nulls removed settings
The concrete resource SHALL override `Update` to compare old and new settings, and explicitly null out removed keys.

#### Scenario: Remove a setting
- **GIVEN** an existing state with setting `x=y`
- **WHEN** the plan removes setting `x`
- **THEN** update SHALL PUT `x=null` to Elasticsearch
- **AND** the flattened state SHALL omit `x`

### Requirement: Delete override reverts all configured settings
The concrete resource SHALL override `Delete` to set all configured settings to null.

#### Scenario: Delete resource
- **GIVEN** an existing state with configured settings
- **WHEN** delete runs
- **THEN** it SHALL PUT all configured keys as null

### Requirement: Read callback returns only configured settings
The `readFunc` callback SHALL GET cluster settings and return only keys that are configured in Terraform.

#### Scenario: Read returns subset
- **GIVEN** a state where persistent had setting `x` configured
- **WHEN** read runs
- **THEN** the returned model SHALL only include `x` if it is still present in Elasticsearch

### Requirement: Settings element type supports string and list values
The nested setting object SHALL support exactly one of `value` (string) or `value_list` (list of strings).

#### Scenario: String value
- **GIVEN** a setting with `value = "foo"`
- **WHEN** the setting is expanded
- **THEN** the API map SHALL contain `"foo"`

#### Scenario: List value
- **GIVEN** a setting with `value_list = ["a", "b"]`
- **WHEN** the setting is expanded
- **THEN** the API map SHALL contain `["a", "b"]`
