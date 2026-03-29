# `elasticstack_elasticsearch_index_lifecycle` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/ilm.go`

## Purpose

Manage Elasticsearch Index Lifecycle Management (ILM) policies. Each policy defines a sequence of lifecycle phases (`hot`, `warm`, `cold`, `frozen`, `delete`) and the actions to execute in each phase. The resource creates or updates the policy on create/update and removes it on destroy.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_lifecycle" "example" {
  id            = <computed, string>  # internal identifier: <cluster_uuid>/<policy_name>
  name          = <required, string>  # force new; ILM policy name
  metadata      = <optional, string>  # JSON string; diff-suppressed (semantic equality)
  modified_date = <computed, string>  # datetime of last modification from API

  # At least one of: hot, warm, cold, frozen, delete

  hot {  # optional; max 1
    min_age = <optional+computed, string>  # minimum age to enter phase

    set_priority {  # optional; max 1
      priority = <required, int>  # >= 0
    }
    unfollow {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    rollover {  # optional; max 1
      max_age                = <optional, string>
      max_docs               = <optional, int>
      max_size               = <optional, string>
      max_primary_shard_docs = <optional, int>   # ES >= 8.2
      max_primary_shard_size = <optional, string>
      min_age                = <optional, string>                  # ES >= 8.4
      min_docs               = <optional, int>                     # ES >= 8.4
      min_size               = <optional, string>                  # ES >= 8.4
      min_primary_shard_docs = <optional, int>                     # ES >= 8.4
      min_primary_shard_size = <optional, string>                  # ES >= 8.4
    }
    readonly {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    shrink {  # optional; max 1
      number_of_shards         = <optional, int>
      max_primary_shard_size   = <optional, string>
      allow_write_after_shrink = <optional, bool>   # ES >= 8.14
    }
    forcemerge {  # optional; max 1
      max_num_segments = <required, int>  # >= 1
      index_codec      = <optional, string>
    }
    searchable_snapshot {  # optional; max 1
      snapshot_repository = <required, string>
      force_merge_index   = <optional, bool>  # default true
    }
    downsample {  # optional; max 1
      fixed_interval = <required, string>
      wait_timeout   = <optional+computed, string>
    }
  }

  warm {  # optional; max 1
    min_age = <optional+computed, string>

    set_priority {  # optional; max 1
      priority = <required, int>  # >= 0
    }
    unfollow {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    readonly {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    allocate {  # optional; max 1
      number_of_replicas   = <optional, int>    # default 0
      total_shards_per_node = <optional, int>   # default -1; ES >= 7.16
      include              = <optional, string>  # JSON object; default "{}"
      exclude              = <optional, string>  # JSON object; default "{}"
      require              = <optional, string>  # JSON object; default "{}"
    }
    migrate {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    shrink {  # optional; max 1
      number_of_shards         = <optional, int>
      max_primary_shard_size   = <optional, string>
      allow_write_after_shrink = <optional, bool>   # ES >= 8.14
    }
    forcemerge {  # optional; max 1
      max_num_segments = <required, int>  # >= 1
      index_codec      = <optional, string>
    }
    downsample {  # optional; max 1
      fixed_interval = <required, string>
      wait_timeout   = <optional+computed, string>
    }
  }

  cold {  # optional; max 1
    min_age = <optional+computed, string>

    set_priority {  # optional; max 1
      priority = <required, int>  # >= 0
    }
    unfollow {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    readonly {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    searchable_snapshot {  # optional; max 1
      snapshot_repository = <required, string>
      force_merge_index   = <optional, bool>  # default true
    }
    allocate {  # optional; max 1
      number_of_replicas    = <optional, int>    # default 0
      total_shards_per_node = <optional, int>    # default -1; ES >= 7.16
      include               = <optional, string>  # JSON object; default "{}"
      exclude               = <optional, string>  # JSON object; default "{}"
      require               = <optional, string>  # JSON object; default "{}"
    }
    migrate {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    freeze {  # optional; max 1
      enabled = <optional, bool>  # default true
    }
    downsample {  # optional; max 1
      fixed_interval = <required, string>
      wait_timeout   = <optional+computed, string>
    }
  }

  frozen {  # optional; max 1
    min_age = <optional+computed, string>

    searchable_snapshot {  # optional; max 1
      snapshot_repository = <required, string>
      force_merge_index   = <optional, bool>  # default true
    }
  }

  delete {  # optional; max 1
    min_age = <optional+computed, string>

    wait_for_snapshot {  # optional; max 1
      policy = <required, string>  # SLM policy name
    }
    delete {  # optional; max 1
      delete_searchable_snapshot = <optional, bool>  # default true
    }
  }

  elasticsearch_connection {  # optional; deprecated
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

### Requirement: ILM policy CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Lifecycle API (`ILM.PutLifecycle`) to create and update ILM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html)). The resource SHALL use the Elasticsearch Get Lifecycle API (`ILM.GetLifecycle`) to read ILM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-get-lifecycle.html)). The resource SHALL use the Elasticsearch Delete Lifecycle API (`ILM.DeleteLifecycle`) to delete ILM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-delete-lifecycle.html)). When Elasticsearch returns a non-success response for any create, update, read, or delete request (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create/update

- GIVEN the Put Lifecycle API returns a non-success response
- WHEN create or update runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: API failure on delete

- GIVEN the Delete Lifecycle API returns a non-success response
- WHEN delete runs
- THEN Terraform diagnostics SHALL include the API error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`. During create and update, the resource SHALL compute `id` from the cluster UUID (via `client.ID`) and the configured `name`.

#### Scenario: ID set after create

- GIVEN a successful create
- WHEN create completes
- THEN state SHALL contain `id` in the form `<cluster_uuid>/<policy_name>`

### Requirement: Import (REQ-007)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. Import SHALL accept an `id` in the format `<cluster_uuid>/<policy_name>`; subsequent read and delete operations SHALL parse this composite `id` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN a valid composite id `<cluster_uuid>/<policy_name>`
- WHEN import runs
- THEN the id SHALL be stored and subsequent read SHALL populate all policy attributes

### Requirement: Lifecycle (REQ-008)

Changing `name` SHALL require replacement (`ForceNew`); an in-place rename is not supported.

#### Scenario: Name change triggers replacement

- GIVEN an existing ILM policy
- WHEN `name` is changed in configuration
- THEN Terraform SHALL plan a replacement (destroy + create)

### Requirement: Phase requirement (REQ-009)

At least one of `hot`, `warm`, `cold`, `frozen`, or `delete` phases SHALL be configured. If none is specified, the schema SHALL reject the configuration before any API call.

#### Scenario: No phase configured

- GIVEN a configuration with none of the five phase blocks
- WHEN plan or apply runs
- THEN the provider SHALL return a validation error

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, update, read, delete).

#### Scenario: Resource-level connection

- GIVEN `elasticsearch_connection` is set with custom endpoints
- WHEN API calls run
- THEN they SHALL use the resource-scoped client, not the provider client

### Requirement: Create and update (REQ-012–REQ-013)

On create and update, the resource SHALL expand the Terraform configuration into a `models.Policy` struct, set `policy.Name` from `name`, and submit it using the Put Lifecycle API. After a successful put, the resource SHALL set `id` and perform a read-after-write to refresh all computed attributes in state.

#### Scenario: Read-after-write on create

- GIVEN a successful Put Lifecycle call
- WHEN create completes
- THEN state SHALL reflect all computed attributes (e.g. `modified_date`) populated by the subsequent read

### Requirement: Read (REQ-014–REQ-016)

On read, the resource SHALL parse `id` using `clients.CompositeIDFromStr` to extract the policy name, call the Get Lifecycle API, and remove the resource from state when the policy is not found (HTTP 404). If the Get Lifecycle API succeeds but does not include the named policy in its response, the resource SHALL return an error diagnostic. On read, the resource SHALL set `modified_date`, `name`, and (when present) `metadata` from the API response, and SHALL flatten all returned phase configurations into the corresponding phase blocks in state.

#### Scenario: Policy not found on read

- GIVEN the ILM policy has been deleted out-of-band
- WHEN read runs
- THEN the resource SHALL be removed from state (id cleared) without an error

#### Scenario: Policy missing from response body

- GIVEN Get Lifecycle returns a success response but the named policy is absent from the response map
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Delete (REQ-017)

On delete, the resource SHALL parse `id` using `clients.CompositeIDFromStr` to extract the policy name and call the Delete Lifecycle API with that name.

#### Scenario: Delete by policy name

- GIVEN an existing ILM policy with a known composite id
- WHEN delete runs
- THEN the provider SHALL call Delete Lifecycle with the policy name extracted from `id`

### Requirement: Metadata mapping (REQ-018–REQ-019)

`metadata` SHALL be validated as JSON by the schema (`validation.StringIsJSON`) and SHALL have JSON semantic diff suppression applied (`DiffJSONSuppress`). On create/update, when `metadata` is set, it SHALL be decoded into a `map[string]any` and included in the policy request body; if JSON decoding fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. On read, when the API response includes policy metadata, it SHALL be serialized to a JSON string and stored in state.

#### Scenario: Invalid metadata JSON

- GIVEN `metadata` is set to a non-JSON string
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic before calling Put Lifecycle

### Requirement: Phase and action mapping (REQ-020–REQ-024)

On create/update, for each configured phase the resource SHALL expand the phase's `min_age` and each action block into the API model. Actions with an `enabled` field (`readonly`, `freeze`, `unfollow`) SHALL be omitted from the API request when `enabled` is false. The `allocate` action's `include`, `exclude`, and `require` fields SHALL be validated as JSON objects by the schema and SHALL be decoded from JSON strings into `map[string]any` before inclusion in the API request. On read, the resource SHALL flatten phase actions returned by the API back into their corresponding Terraform block structure, including re-serializing `allocate` allocation filter fields (`include`, `exclude`, `require`) to JSON strings.

#### Scenario: readonly action disabled

- GIVEN `hot.0.readonly.0.enabled` is false
- WHEN create or update runs
- THEN the `readonly` key SHALL be absent from the `hot` phase actions in the API request body

#### Scenario: allocate filter round-trip

- GIVEN `warm.0.allocate.0.exclude` is set to `{"box_type":"hot"}`
- WHEN create and subsequent read run
- THEN state SHALL contain `exclude` equal to `{"box_type":"hot"}`

### Requirement: Compatibility — rollover min conditions (REQ-025)

When any of `rollover.min_age`, `rollover.min_docs`, `rollover.min_size`, `rollover.min_primary_shard_docs`, or `rollover.min_primary_shard_size` is set to a non-default value, the resource SHALL verify the Elasticsearch server version is at least **8.4.0**. If the server version is lower and the value differs from the default, the resource SHALL fail with an error diagnostic and SHALL not call the Put Lifecycle API.

#### Scenario: min rollover condition on unsupported version

- GIVEN server version < 8.4.0 and `hot.0.rollover.0.min_age` is set to a non-empty value
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic referencing the unsupported setting

### Requirement: Compatibility — max_primary_shard_docs (REQ-026)

When `rollover.max_primary_shard_docs` is set to a non-zero value, the resource SHALL verify the Elasticsearch server version is at least **8.2.0**. If the server version is lower and the value differs from the default (`0`), the resource SHALL fail with an error diagnostic.

#### Scenario: max_primary_shard_docs on unsupported version

- GIVEN server version < 8.2.0 and `hot.0.rollover.0.max_primary_shard_docs` is set to a non-zero value
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Compatibility — total_shards_per_node (REQ-027)

When `allocate.total_shards_per_node` is set to a value other than `-1`, the resource SHALL verify the Elasticsearch server version is at least **7.16.0**. If the server version is lower and the value differs from the default (`-1`), the resource SHALL fail with an error diagnostic.

#### Scenario: total_shards_per_node on unsupported version

- GIVEN server version < 7.16.0 and `warm.0.allocate.0.total_shards_per_node` is set to a value other than -1
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Compatibility — allow_write_after_shrink (REQ-028)

When `shrink.allow_write_after_shrink` is set to `true`, the resource SHALL verify the Elasticsearch server version is at least **8.14.0**. If the server version is lower and the value differs from the default (`false`), the resource SHALL fail with an error diagnostic.

#### Scenario: allow_write_after_shrink on unsupported version

- GIVEN server version < 8.14.0 and `hot.0.shrink.0.allow_write_after_shrink` is true
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic

### Requirement: State — boolean-action read preservation (REQ-029)

For the `readonly`, `freeze`, and `unfollow` actions, when the API response does not include the action (because it was not active), the resource SHALL check the prior Terraform configuration. If the prior configuration contained the action block, the resource SHALL write `enabled = false` into state for that block. This ensures that an explicit `enabled = false` in config does not cause an endless diff loop.

#### Scenario: readonly absent from API response but present in config

- GIVEN prior config contained `hot.0.readonly.0.enabled = false`
- WHEN read runs and the API response omits `readonly` from the hot phase
- THEN state SHALL contain `hot.0.readonly.0.enabled = false`

### Requirement: State — allocate total_shards_per_node default (REQ-030)

On read, when the API response for an `allocate` action does not include `total_shards_per_node`, the resource SHALL store `-1` (the default) in state to avoid perpetual plan drift on Elasticsearch versions that do not return this field.

#### Scenario: total_shards_per_node absent from API

- GIVEN an `allocate` action in the API response that omits `total_shards_per_node`
- WHEN read runs
- THEN state SHALL contain `total_shards_per_node = -1`
