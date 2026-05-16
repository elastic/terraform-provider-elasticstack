# `elasticstack_elasticsearch_cluster_settings` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/cluster/settings/`

## Purpose

Manage cluster-wide settings in Elasticsearch via the Cluster Update Settings API. The resource supports both persistent settings (survive a full cluster restart) and transient settings (reset on cluster restart), using a flat-settings representation for each category. Each tracked setting is identified by name and may carry either a scalar value or a list of values.

## Schema

```hcl
resource "elasticstack_elasticsearch_cluster_settings" "example" {
  id = <computed, string> # internal identifier: <cluster_uuid>/cluster-settings

  persistent {           # optional, SingleNestedBlock
    setting {            # required, SetNestedBlock, min 1 item
      name       = <required, string>       # setting key
      value      = <optional, string>       # scalar value (mutually exclusive with value_list)
      value_list = <optional, list(string)> # list value (mutually exclusive with value)
    }
  }

  transient {            # optional, SingleNestedBlock
    setting {            # required, SetNestedBlock, min 1 item
      name       = <required, string>
      value      = <optional, string>
      value_list = <optional, list(string)>
    }
  }

  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }
}
```
## Requirements
### Requirement: Cluster settings APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Cluster Update Settings API to create, update, and delete cluster settings ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-update-settings.html)). The resource SHALL use the Elasticsearch Cluster Get Settings API with flat settings enabled (`flat_settings=true`) to read the current cluster settings. When Elasticsearch returns a non-success response for any API call, the resource SHALL surface the error to Terraform diagnostics and stop processing.

#### Scenario: API failure on put

- GIVEN the Cluster Update Settings API returns a non-success response
- WHEN create, update, or delete runs
- THEN Terraform diagnostics SHALL include the error and no state update SHALL occur

#### Scenario: API failure on read

- GIVEN the Cluster Get Settings API returns a non-success response
- WHEN read runs
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity (REQ-004–REQ-005)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/cluster-settings`. During create, the resource SHALL derive the `id` by obtaining the cluster UUID from the API client and appending the fixed suffix `cluster-settings`.

#### Scenario: ID set on create

- GIVEN a successful create
- WHEN the resource is created
- THEN `id` SHALL be set to `<cluster_uuid>/cluster-settings` in Terraform state

### Requirement: Import (REQ-006)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. After import, the `id` SHALL end with `/cluster-settings`.

#### Scenario: Import passthrough

- GIVEN import with a valid composite id ending in `/cluster-settings`
- WHEN import completes
- THEN the `id` SHALL be stored in state and subsequent read SHALL refresh settings

### Requirement: Connection (REQ-007–REQ-008)

By default, the resource SHALL use the provider-level Elasticsearch client for all API calls. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls, overriding the provider-level client.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN API calls run (create, read, update, delete)
- THEN they SHALL use the resource-scoped client derived from `elasticsearch_connection`

### Requirement: Create and update (REQ-009–REQ-011)

On create and update, the envelope SHALL invoke `WriteFunc[tfModel]` callbacks wired into the `Create` and `Update` slots of `ElasticsearchResourceOptions[tfModel]`. Each callback SHALL receive `WriteRequest[tfModel]` (`Plan`, `Prior`, `Config`, `WriteID`); on create `req.Prior` is `nil`, and on update `req.Prior` is a non-nil pointer to the prior-state model. Each callback SHALL expand the configured `persistent` and `transient` blocks into a flat settings map and submit them to the Cluster Update Settings API. When the configuration is updated and a setting present in the previous state is absent from the new plan, the update callback SHALL include that setting name with a `null` value in the API request to explicitly remove it from the cluster. After a successful put, the envelope SHALL perform read-after-write via the shared read callback to refresh state (including composite `id`).

#### Scenario: Setting removed on update

- GIVEN a setting was present in `persistent` (or `transient`) in the previous state
- AND that setting is absent from the updated configuration
- WHEN update runs
- THEN the resource SHALL send the setting with value `null` to unset it in the cluster

#### Scenario: Read-after-write

- GIVEN a successful put (create or update)
- WHEN the resource finishes writing
- THEN the resource SHALL call the Cluster Get Settings API and update state from the response

### Requirement: Read (REQ-012–REQ-013)

On read, the resource SHALL call the Cluster Get Settings API and, for each configured setting name tracked in state, SHALL update the corresponding state attribute from the API response. Settings that are no longer present in the API response SHALL be dropped from state during read. The resource SHALL NOT remove itself from state on read because cluster settings are global and always present.

#### Scenario: Setting dropped from read response

- GIVEN a setting key is tracked in state
- AND the Cluster Get Settings API response does not contain that key
- WHEN read runs
- THEN that setting SHALL be absent from state after read

### Requirement: Delete (REQ-014)

On delete, the resource SHALL set all tracked `persistent` and `transient` setting names to `null` in the Cluster Update Settings API request, effectively removing those settings from the cluster. The resource SHALL derive the list of settings to remove from the current state rather than contacting the cluster to discover them.

#### Scenario: Delete clears all tracked settings

- GIVEN a resource with persistent and transient settings
- WHEN delete runs
- THEN all tracked setting names SHALL be sent with `null` value in the Cluster Update Settings API call

### Requirement: Setting value validation (REQ-015–REQ-017)

Each `setting` block within `persistent` or `transient` MUST specify exactly one of `value` (non-empty string) or `value_list` (non-empty list). If both `value` and `value_list` are non-empty, the resource SHALL return an error diagnostic before calling any API. If neither `value` nor `value_list` is provided with a non-empty value, the resource SHALL return an error diagnostic before calling any API. Setting names MUST be unique within a `persistent` or `transient` block; if a duplicate name is found, the resource SHALL return an error diagnostic before calling any API.

#### Scenario: Both value and value_list set

- GIVEN a setting with both `value` and `value_list` non-empty
- WHEN the configuration is validated (plan or apply)
- THEN the provider SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

#### Scenario: Neither value nor value_list set

- GIVEN a setting with an empty `value` and an empty `value_list`
- WHEN the configuration is validated (plan or apply)
- THEN the provider SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

#### Scenario: Duplicate setting names

- GIVEN two `setting` blocks with the same `name` within one `persistent` or `transient` block
- WHEN the configuration is validated (plan or apply)
- THEN the provider SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

### Requirement: State mapping (REQ-018–REQ-019)

On read, for each setting tracked in state, the resource SHALL read the value from the flat API response and store it as either `value` (if the API returns a string) or `value_list` (if the API returns a list) in the corresponding state attribute. Settings returned by the API but not tracked in state SHALL NOT be written to state, so only explicitly managed settings are reflected in Terraform state.

#### Scenario: Scalar value read back

- GIVEN a persistent setting with a string value in the API response
- WHEN read runs
- THEN `setting.value` SHALL be set to that string in state

#### Scenario: List value read back

- GIVEN a transient setting with a list value in the API response
- WHEN read runs
- THEN `setting.value_list` SHALL be set to that list in state

### Requirement: Typed client implementation for cluster settings
The resource SHALL use the go-elasticsearch Typed API for cluster settings operations. `GetSettings` SHALL use `Cluster.GetSettings().Do(ctx)` with flat settings enabled. `PutSettings` SHALL use `Cluster.PutSettings().Do(ctx)`. The typed response maps (`Persistent`, `Transient`, `Defaults`) are unmarshaled via `json.RawMessage` to maintain the existing `map[string]any` contract; manual JSON decoding of raw HTTP response bodies SHALL NOT occur.

#### Scenario: Typed API read with flat settings
- GIVEN a successful Cluster Get Settings API call
- WHEN the provider processes the response
- THEN the typed API `getsettings.Response` SHALL provide `Persistent`, `Transient`, and `Defaults` as `map[string]json.RawMessage`
- AND the provider SHALL unmarshal each `RawMessage` value to `any` to maintain the existing `map[string]any` contract with callers

#### Scenario: Typed API write sends settings
- GIVEN cluster settings to update
- WHEN the provider calls the Cluster Put Settings API
- THEN the request SHALL be built using typed API request builders
- AND manual `json.Marshal` of a `map[string]any` into a raw request body SHALL NOT occur

### Requirement: Resource is implemented in Plugin Framework

The `elasticstack_elasticsearch_cluster_settings` resource SHALL be implemented using the Terraform Plugin Framework instead of the Plugin SDK. It SHALL embed `*entitycore.ElasticsearchResource[tfModel]` and satisfy `resource.Resource`, `resource.ResourceWithConfigure`, `resource.ResourceWithImportState`, `resource.ResourceWithUpgradeState`, and `resource.ResourceWithValidateConfig`.

#### Scenario: Provider registrar uses PF resource

- **WHEN** the provider builds its resource map
- **THEN** `elasticstack_elasticsearch_cluster_settings` SHALL be a Plugin Framework resource
- **AND** the SDK version SHALL no longer be registered

### Requirement: Model satisfies envelope constraint

The cluster-settings model SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model type assertion

- **WHEN** `entitycore.NewElasticsearchResource[tfModel]("cluster_settings", opts)` is constructed with the cluster-settings schema and lifecycle callbacks
- **THEN** compilation SHALL succeed

### Requirement: Schema factory returns blocks without connection injection

The schema factory SHALL return a `schema.Schema` with `persistent` and `transient` as `SingleNestedBlock`, each containing a `SetNestedBlock` named `setting` (min 1). Each `setting` SHALL have `name` (required string), `value` (optional string), and `value_list` (optional list of strings). The factory SHALL NOT include the `elasticsearch_connection` block.

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

### Requirement: Envelope create callback builds flat settings map and PUTs

The envelope create callback SHALL expand `persistent` and `transient` from `WriteRequest[tfModel].Plan` (with `req.Prior == nil`) into flat settings maps and call the Cluster Update Settings API. It SHALL derive the composite ID on the returned model before returning `WriteResult[tfModel]`. Validation (duplicate names, value/value_list exclusivity, and non-empty category block) is enforced at plan time by schema validators and `ValidateConfig`.

#### Scenario: Create with persistent and transient settings

- **GIVEN** a plan with both `persistent` and `transient` settings
- **WHEN** create runs
- **THEN** the Cluster Update Settings API SHALL receive both categories
- **AND** the envelope read-after-write SHALL persist `id` as `<cluster_uuid>/cluster-settings`

### Requirement: Envelope update callback nulls out removed settings

The envelope update callback SHALL decode prior settings from `*WriteRequest[tfModel].Prior` (which is non-nil on update) and compare setting names against `Plan`. For each of `persistent` and `transient`, any setting name present in the prior model but absent from the new plan SHALL be included in the PUT request with a `null` value so Elasticsearch removes it.

#### Scenario: Setting removed from configuration

- **GIVEN** the old state contained `persistent` setting `cluster.max_shards_per_node`
- **AND** the new plan does not contain that setting
- **WHEN** update runs
- **THEN** the PUT request SHALL include `"cluster.max_shards_per_node": null`

### Requirement: Envelope delete callback removes all tracked settings

The envelope delete callback SHALL derive all tracked setting names from the current state for both `persistent` and `transient`. It SHALL call the Cluster Update Settings API with each tracked name set to `null`.

#### Scenario: Delete clears tracked settings

- **GIVEN** a resource tracking three persistent settings and two transient settings
- **WHEN** delete runs
- **THEN** all five names SHALL be sent with `null` values

### Requirement: Validation preserved

Each `setting` block SHALL enforce that exactly one of `value` or `value_list` is non-empty. Duplicate `name` values within a single `persistent` or `transient` block SHALL produce an error diagnostic before any API call.

#### Scenario: Invalid setting validation

- **GIVEN** a `setting` with both `value` and `value_list` set
- **WHEN** the configuration is validated (plan or apply)
- **THEN** the provider SHALL return an error diagnostic

### Requirement: Non-empty configuration validation

The resource SHALL enforce that at least one of `persistent` or `transient` contains at least one `setting` block. An empty or null configuration for both categories SHALL produce an error diagnostic at plan time.

#### Scenario: Both persistent and transient empty

- **GIVEN** a resource with no `setting` blocks in `persistent` and `transient`
- **WHEN** the configuration is validated (plan or apply)
- **THEN** the provider SHALL return an error diagnostic

### Requirement: State upgrade

The resource SHALL implement `resource.ResourceWithUpgradeState` to migrate state written by the SDKv2-based implementation (schema version 0) to the Plugin Framework implementation (schema version 1). The upgrader SHALL unwrap list-of-one category blocks into `SingleNestedBlock` objects and normalise empty-string/`[]` values for the unused `value`/`value_list` alternative to null.

#### Scenario: SDKv2 state upgrade

- **GIVEN** Terraform state written by schema version 0
- **WHEN** the resource is refreshed
- **THEN** the state SHALL be transparently upgraded to schema version 1

### Requirement: Import preserved

The resource SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import

- **GIVEN** an import id of `<cluster_uuid>/cluster-settings`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state
- **AND** subsequent read SHALL refresh settings

