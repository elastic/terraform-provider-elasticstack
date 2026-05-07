## ADDED Requirements

### Requirement: Resource is implemented in Plugin Framework

The `elasticstack_elasticsearch_cluster_settings` resource SHALL be implemented using the Terraform Plugin Framework instead of the Plugin SDK. It SHALL embed `*entitycore.ElasticsearchResource[tfModel]` and satisfy `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.

#### Scenario: Provider registrar uses PF resource

- **WHEN** the provider builds its resource map
- **THEN** `elasticstack_elasticsearch_cluster_settings` SHALL be a Plugin Framework resource
- **AND** the SDK version SHALL no longer be registered

### Requirement: Model satisfies envelope constraint

The cluster-settings model SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model type assertion

- **WHEN** `entitycore.NewElasticsearchResource[Data]` is called with the cluster-settings model
- **THEN** compilation SHALL succeed

### Requirement: Schema factory returns blocks without connection injection

The schema factory SHALL return a `schema.Schema` with `persistent` and `transient` as `ListNestedBlock` (max 1), each containing a `SetNestedAttribute` named `setting` (min 1). Each `setting` SHALL have `name` (required string), `value` (optional string), and `value_list` (optional list of strings). The factory SHALL NOT include the `elasticsearch_connection` block.

#### Scenario: Schema shape preserved

- **WHEN** the envelope's `Schema` method returns the final schema
- **THEN** it SHALL contain `persistent` and `transient` blocks
- **AND** each SHALL contain a set of `setting` entries
- **AND** it SHALL contain the injected `elasticsearch_connection` block

### Requirement: Read callback populates settings from flat API response

The read callback SHALL call Elasticsearch Cluster Get Settings API with `flat_settings=true`. For each setting name tracked in Terraform state under `persistent` or `transient`, it SHALL read the corresponding flat key from the API response and store it as either `value` (if the API returns a string) or `value_list` (if the API returns a list). Settings not present in the API response SHALL be omitted from the corresponding state block.

#### Scenario: Scalar setting read back

- **GIVEN** a flat settings response containing `"persistent": {"indices.recovery.max_bytes_per_sec": "40mb"}`
- **WHEN** read runs for a resource tracking that key
- **THEN** the state SHALL contain `value = "40mb"` for that setting

#### Scenario: List setting read back

- **GIVEN** a flat settings response containing `"persistent": {"search.remote.connect": ["true", "false"]}`
- **WHEN** read runs for a resource tracking that key
- **THEN** the state SHALL contain `value_list = ["true", "false"]` for that setting

### Requirement: Create override builds flat settings map and PUTs

The concrete `Create` method SHALL decode the plan model, expand `persistent` and `transient` blocks into flat settings maps, validate settings (duplicate names, value/value_list exclusivity), and call the Cluster Update Settings API. After a successful PUT, it SHALL derive the composite ID and persist state via the read callback.

#### Scenario: Create with persistent and transient settings

- **GIVEN** a plan with both `persistent` and `transient` settings
- **WHEN** create runs
- **THEN** the Cluster Update Settings API SHALL receive both categories
- **AND** the resource id SHALL be set to `<cluster_uuid>/cluster-settings`

### Requirement: Update override nulls out removed settings

The concrete `Update` method SHALL decode both plan and state. For each of `persistent` and `transient`, it SHALL compare the old setting names with the new setting names. Any setting name present in the old state but absent from the new plan SHALL be included in the PUT request with a `null` value so Elasticsearch removes it.

#### Scenario: Setting removed from configuration

- **GIVEN** the old state contained `persistent` setting `cluster.max_shards_per_node`
- **AND** the new plan does not contain that setting
- **WHEN** update runs
- **THEN** the PUT request SHALL include `"cluster.max_shards_per_node": null`

### Requirement: Delete override removes all tracked settings

The concrete `Delete` method SHALL derive all tracked setting names from the current state for both `persistent` and `transient`. It SHALL call the Cluster Update Settings API with each tracked name set to `null`.

#### Scenario: Delete clears tracked settings

- **GIVEN** a resource tracking three persistent settings and two transient settings
- **WHEN** delete runs
- **THEN** all five names SHALL be sent with `null` values

### Requirement: Validation preserved

Each `setting` block SHALL enforce that exactly one of `value` or `value_list` is non-empty. Duplicate `name` values within a single `persistent` or `transient` block SHALL produce an error diagnostic before any API call.

#### Scenario: Invalid setting validation

- **GIVEN** a `setting` with both `value` and `value_list` set
- **WHEN** create or update runs
- **THEN** the provider SHALL return an error diagnostic

### Requirement: Import preserved

The resource SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import

- **GIVEN** an import id of `<cluster_uuid>/cluster-settings`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state
- **AND** subsequent read SHALL refresh settings

## MODIFIED Requirements

None — externally-observable behavior is preserved verbatim.

## REMOVED Requirements

None.
