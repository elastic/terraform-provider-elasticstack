# `elasticstack_elasticsearch_cluster_settings` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/cluster/settings.go`

## Purpose

Manage cluster-wide settings in Elasticsearch via the Cluster Update Settings API. The resource supports both persistent settings (survive a full cluster restart) and transient settings (reset on cluster restart), using a flat-settings representation for each category. Each tracked setting is identified by name and may carry either a scalar value or a list of values.

## Schema

```hcl
resource "elasticstack_elasticsearch_cluster_settings" "example" {
  id = <computed, string> # internal identifier: <cluster_uuid>/cluster-settings

  persistent {           # optional, max 1 block
    setting {            # required, set, min 1 item
      name       = <required, string>       # setting key
      value      = <optional, string>       # scalar value (mutually exclusive with value_list)
      value_list = <optional, list(string)> # list value (mutually exclusive with value)
    }
  }

  transient {            # optional, max 1 block
    setting {            # required, set, min 1 item
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

On create and update, the resource SHALL expand the configured `persistent` and `transient` blocks into a flat settings map and submit it to the Cluster Update Settings API. When the configuration is updated and a setting present in the previous state is absent from the new state, the resource SHALL include that setting name with a `null` value in the API request to explicitly remove it from the cluster. After a successful put, the resource SHALL set `id` and perform a read to refresh state.

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
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

#### Scenario: Neither value nor value_list set

- GIVEN a setting with an empty `value` and an empty `value_list`
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

#### Scenario: Duplicate setting names

- GIVEN two `setting` blocks with the same `name` within one `persistent` or `transient` block
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL NOT call the Cluster Update Settings API

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
