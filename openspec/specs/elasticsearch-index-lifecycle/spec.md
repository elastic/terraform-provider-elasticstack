# `elasticstack_elasticsearch_index_lifecycle` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/ilm/`

## Purpose

Define the Terraform schema and runtime behavior for managing **Elasticsearch index lifecycle management (ILM) policies**: creating and updating policies on the cluster, reading them into state (including refresh and drift detection), deleting them, importing existing policies, choosing the Elasticsearch connection, and enforcing **server-version gates** for ILM action fields that only exist on newer Elasticsearch releases.

## Schema

### Top-level attributes

```hcl
resource "elasticstack_elasticsearch_index_lifecycle" "example" {
  id            = <computed, string> # <cluster_uuid>/<policy_name>
  name          = <required, string> # policy identifier; force new
  metadata      = <optional, json string> # valid JSON; normalized diff
  modified_date = <computed, string> # last modification time from the cluster

  # At least one of hot, warm, cold, frozen, delete MUST be set (schema: AtLeastOneOf).
  # Each phase is a Plugin Framework SingleNestedBlock (at most one block per phase; state stores a single object or null).
  hot    { /* phase_hot */ }
  warm   { /* phase_warm */ }
  cold   { /* phase_cold */ }
  frozen { /* phase_frozen */ }
  delete { /* phase_delete */ }

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

In Terraform configuration, each phase is written as a **`SingleNestedBlock`** (for example `hot { ... }`). State stores that phase as an object-shaped value (or null when absent), not as a single-element list.

### Per-phase object (common)

Every phase object MAY include:

| Attribute | Constraint | Notes |
|-----------|--------------|--------|
| `min_age` | optional + computed, string | Minimum age before entering this phase; may be populated from the cluster on read. |

### Allowed nested actions by phase

| Phase | Nested action blocks (each is a **`SingleNestedBlock`**) |
|-------|-----------------------------------------------------------------------------|
| **hot** | `set_priority`, `unfollow`, `rollover`, `readonly`, `shrink`, `forcemerge`, `searchable_snapshot`, `downsample` |
| **warm** | `set_priority`, `unfollow`, `readonly`, `allocate`, `migrate`, `shrink`, `forcemerge`, `downsample` |
| **cold** | `set_priority`, `unfollow`, `readonly`, `searchable_snapshot`, `allocate`, `migrate`, `freeze`, `downsample` |
| **frozen** | `searchable_snapshot` only (plus `min_age`) |
| **delete** | `wait_for_snapshot`, `delete` (the ILM delete action; plus `min_age`) |

### Nested action block schemas

Each action below is expressed as Terraform nested block syntax. All such blocks are **optional** and use **`SingleNestedBlock`** semantics (`action { ... }`); state stores each declared action as an object, not as a list.

```hcl
# allocate — warm, cold only
allocate {
  number_of_replicas     = <optional, int, default 0>
  total_shards_per_node  = <optional, int, default -1> # ES >= 7.16 when non-default
  include                = <optional, json string, default "{}"> # JSON object as string; normalized diff
  exclude                = <optional, json string, default "{}">
  require                = <optional, json string, default "{}">
}

# delete — delete phase only (ILM action that removes the index)
delete {
  delete_searchable_snapshot = <optional, bool, default true>
}

# forcemerge — hot, warm only
# When the block is omitted, max_num_segments is not required. When the block is declared, max_num_segments is required (object-level AlsoRequires).
forcemerge {
  max_num_segments = <optional, int, >= 1> # required when block is present
  index_codec        = <optional, string>
}

# freeze — cold only
freeze {
  enabled = <optional, bool, default true> # when false, action omitted from API (see requirements)
}

# migrate — warm, cold only
migrate {
  enabled = <optional, bool, default true>
}

# readonly — hot, warm, cold only
readonly {
  enabled = <optional, bool, default true>
}

# rollover — hot only
rollover {
  max_age                 = <optional, string>
  max_docs                = <optional, int>
  max_size                = <optional, string>
  max_primary_shard_docs  = <optional, int> # ES >= 8.2 when non-default
  max_primary_shard_size  = <optional, string>
  min_age                 = <optional, string> # ES >= 8.4 when non-default
  min_docs                = <optional, int>  # ES >= 8.4 when non-default
  min_size                = <optional, string>
  min_primary_shard_docs  = <optional, int>  # ES >= 8.4 when non-default
  min_primary_shard_size  = <optional, string> # ES >= 8.4 when non-default
}

# searchable_snapshot — hot, cold, frozen only
# snapshot_repository required when block is present (object-level AlsoRequires).
searchable_snapshot {
  snapshot_repository = <optional, string> # required when block is present
  force_merge_index   = <optional, bool, default true>
}

# set_priority — hot, warm, cold only
# priority required when block is present (object-level AlsoRequires).
set_priority {
  priority = <optional, int, >= 0> # required when block is present; index recovery priority for this phase
}

# shrink — hot, warm only
shrink {
  number_of_shards           = <optional, int>
  max_primary_shard_size     = <optional, string>
  allow_write_after_shrink     = <optional, bool> # ES >= 8.14 when non-default
}

# unfollow — hot, warm, cold only
unfollow {
  enabled = <optional, bool, default true>
}

# wait_for_snapshot — delete phase only
# policy required when block is present (object-level AlsoRequires).
wait_for_snapshot {
  policy = <optional, string> # required when block is present; SLM policy name to wait for
}

# downsample — hot, warm, cold only
# fixed_interval required when block is present (object-level AlsoRequires).
downsample {
  fixed_interval = <optional, string> # required when block is present
  wait_timeout     = <optional + computed, string> # may be set by the cluster on read
}
```

### Example: fully expanded phase shapes (illustrative)

Each phase is one `SingleNestedBlock` (e.g. `hot { min_age = "1h" ... }`).

```hcl
  hot {
    min_age = <optional+computed, string>

    set_priority { priority = <optional int; required when block present> }
    unfollow { enabled = <optional bool> }
    rollover {
      max_age = <optional string>
      # ... all rollover fields per table above
    }
    readonly { enabled = <optional bool> }
    shrink {
      number_of_shards         = <optional int>
      max_primary_shard_size   = <optional string>
      allow_write_after_shrink = <optional bool>
    }
    forcemerge {
      max_num_segments = <optional int; required when block present>
      index_codec      = <optional string>
    }
    searchable_snapshot {
      snapshot_repository = <optional string; required when block present>
      force_merge_index   = <optional bool>
    }
    downsample {
      fixed_interval = <optional string; required when block present>
      wait_timeout   = <optional+computed string>
    }
  }

  warm {
    min_age = <optional+computed, string>
    set_priority { ... }
    unfollow { ... }
    readonly { ... }
    allocate { ... }
    migrate { ... }
    shrink { ... }
    forcemerge { ... }
    downsample { ... }
  }

  cold {
    min_age = <optional+computed, string>
    set_priority { ... }
    unfollow { ... }
    readonly { ... }
    searchable_snapshot { ... }
    allocate { ... }
    migrate { ... }
    freeze { ... }
    downsample { ... }
  }

  frozen {
    min_age = <optional+computed, string>
    searchable_snapshot {
      snapshot_repository = <optional string; required when block present>
      force_merge_index   = <optional bool>
    }
  }

  delete {
    min_age = <optional+computed, string>
    wait_for_snapshot { policy = <optional string; required when block present> }
    delete { delete_searchable_snapshot = <optional bool> }
  }
```

## Requirements

### Requirement: ILM policy CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch **Put lifecycle policy** API to create and update ILM policies ([Put lifecycle API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html)). The resource SHALL use the **Get lifecycle policy** API to read policies ([Get lifecycle API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-get-lifecycle.html)). The resource SHALL use the **Delete lifecycle policy** API to delete policies ([Delete lifecycle API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-delete-lifecycle.html)).

#### Scenario: Documented APIs for lifecycle operations

- GIVEN an ILM policy managed by this resource
- WHEN create, update, read, or delete runs
- THEN the provider SHALL call the Put, Get, and Delete lifecycle APIs as documented

### Requirement: API error surfacing (REQ-004)

When Elasticsearch returns a non-success response for create, update, or delete, or for read when the response is not a successful retrieval (excluding **not found** on read as specified elsewhere), the resource SHALL surface the error in Terraform diagnostics.

#### Scenario: Non-success response

- GIVEN an Elasticsearch error on create, update, read (other than not found), or delete
- WHEN the provider handles the response
- THEN the error SHALL appear in diagnostics

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`. After a successful create or update, the resource SHALL set `id` using the target cluster identifier and the configured policy `name`.

#### Scenario: Computed id after apply

- GIVEN a successful create or update
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<policy_name>` for the connected cluster and configured name

### Requirement: Import (REQ-007)

The resource SHALL support import using **passthrough** of the stored `id` (no custom import logic). The imported value SHALL be the same composite `id` format used in state.

#### Scenario: Import by id

- GIVEN an import id in the form `<cluster_uuid>/<policy_name>`
- WHEN import completes
- THEN state SHALL retain that `id` for subsequent read

### Requirement: Policy name lifecycle (REQ-008)

When the `name` argument changes, the resource SHALL require **replacement** (new policy identity), not an in-place rename via the same resource instance.

#### Scenario: Renaming a policy

- GIVEN a planned change to `name`
- WHEN Terraform evaluates the resource
- THEN replacement SHALL be required

### Requirement: Elasticsearch connection (REQ-009–REQ-010)

By default, the resource SHALL use the provider-configured Elasticsearch client. When `elasticsearch_connection` is set, the resource SHALL use a **resource-scoped** Elasticsearch client for all API calls for that instance.

#### Scenario: Override connection

- GIVEN `elasticsearch_connection` is configured
- WHEN create, read, update, or delete runs
- THEN API calls SHALL use the connection defined by that block

### Requirement: Create and update flow (REQ-011)

Create and update SHALL both **put** the full policy definition derived from configuration, then **read** the policy back into state so computed fields and cluster-returned values are refreshed.

#### Scenario: Read after write

- GIVEN a successful put
- WHEN create or update finishes
- THEN read logic SHALL run to populate state

### Requirement: Read and absent policy (REQ-012–REQ-013)

Read SHALL parse `id` as a composite identifier; if the format is invalid, the resource SHALL return an error diagnostic. When the lifecycle API indicates the policy **does not exist**, the resource SHALL **remove the resource from state** (empty `id`) and SHALL log a warning that the policy was not found.

#### Scenario: Policy removed outside Terraform

- GIVEN the policy was deleted on the cluster
- WHEN refresh runs
- THEN the resource SHALL be removed from state and SHALL not fail with a hard error solely due to absence

#### Scenario: Invalid stored id

- GIVEN `id` is not `<cluster_uuid>/<resource identifier>`
- WHEN read or delete parses `id`
- THEN the provider SHALL return an error diagnostic describing the required format

### Requirement: Delete (REQ-014)

Delete SHALL derive the policy name from the composite `id` and SHALL call the delete lifecycle API for that name.

#### Scenario: Delete uses policy name from id

- GIVEN a valid `id` in state
- WHEN delete runs
- THEN the delete API SHALL be invoked for the policy name portion of `id`

### Requirement: Phase and metadata validation (REQ-015–REQ-016)

The resource SHALL require **at least one** of the phase blocks `hot`, `warm`, `cold`, `frozen`, or `delete`. The resource SHALL accept `metadata` only if it is **valid JSON**. JSON-valued allocation attributes (`include`, `exclude`, `require`) SHALL be valid JSON; where used, the provider SHALL apply **JSON-aware diff suppression** so equivalent objects do not churn the plan solely due to formatting.

#### Scenario: No phase defined

- GIVEN none of the phase blocks are set
- WHEN Terraform validates configuration
- THEN validation SHALL fail (at least one phase required)

### Requirement: Server version compatibility for optional ILM fields (REQ-017)

For ILM action settings that are only supported starting at a **minimum Elasticsearch version**, the resource SHALL compare the **connected server version** to that minimum when expanding configuration into the API model. If the server is **older** than the required version and the user has set a **non-default** value for that setting, the resource SHALL fail with a diagnostic that instructs removal of the setting or use of the default. If the value equals the default, the resource SHALL **omit** sending that unsupported setting in the policy payload.

#### Scenario: Rollover min conditions on old cluster

- GIVEN Elasticsearch < 8.4 and rollover **min_**\* conditions are set to non-default values
- WHEN create or update expands the policy
- THEN the provider SHALL return an error diagnostic

#### Scenario: Allocate total_shards_per_node on old cluster

- GIVEN Elasticsearch < 7.16 and `total_shards_per_node` is set to a non-default value
- WHEN create or update expands the allocate action
- THEN the provider SHALL return an error diagnostic

### Requirement: Mapping for togglable actions (REQ-018)

For actions **readonly**, **freeze**, and **unfollow**, the resource SHALL send the action to Elasticsearch only when **`enabled` is true**. When **`enabled` is false** but the user still declares the block (so Terraform can express “disabled”), read/flatten SHALL map state in a way that preserves that intent without falsely implying the action is active.

#### Scenario: Disabled readonly block retained in config

- GIVEN the user sets `readonly { enabled = false }` in a phase
- WHEN state is refreshed from the API
- THEN configuration SHALL be able to represent the disabled case without spurious enabled=true drift (per provider flatten rules)

### Requirement: Unknown phase actions (REQ-019)

If expansion encounters an action key that is not supported by the provider’s mapping for that phase, the resource SHALL fail with an error diagnostic indicating the action is not supported.

#### Scenario: Unexpected action in expanded phase map

- GIVEN an internal expansion path surfaces an unknown action name
- WHEN the policy is expanded
- THEN the provider SHALL return a diagnostic

### Requirement: Single nested blocks for phases and actions (REQ-020)

The resource SHALL model each of the phase blocks `hot`, `warm`, `cold`, `frozen`, and `delete` as a **Plugin Framework `SingleNestedBlock`** (at most one block per phase in configuration; state stores a single nested object or null when absent), not as a list nested block with a maximum length of one.

Each ILM action block allowed under a phase (for example `set_priority`, `rollover`, `forcemerge`, `searchable_snapshot`, `wait_for_snapshot`, `delete`, and other actions defined by the provider schema) SHALL likewise be modeled as a **`SingleNestedBlock`**.

The **`elasticsearch_connection`** block SHALL remain a **list nested block** as provided by the shared provider connection schema.

#### Scenario: Phase block cardinality

- GIVEN a Terraform configuration for this resource
- WHEN the user declares a phase (for example `hot { ... }`)
- THEN the schema SHALL allow at most one such block for that phase and SHALL persist that phase as an object-shaped value in state, not as a single-element list

#### Scenario: Action block cardinality

- GIVEN a phase that supports an ILM action block
- WHEN the user declares that action (for example `forcemerge { ... }`)
- THEN the schema SHALL allow at most one such block and SHALL persist it as an object-shaped value in state, not as a single-element list

### Requirement: State schema version and upgrade (REQ-021)

The resource SHALL use a **non-zero** `schema.Schema.Version` for this resource type after this change.

The resource SHALL implement **`ResourceWithUpgradeState`** and SHALL migrate stored Terraform state from the **prior version** (list-shaped nested values for phases and ILM actions) to the **new version** (object-shaped nested values) for the same logical configuration.

The migration SHALL unwrap list-encoded values **only** for known ILM phase keys and known ILM action keys under those phases (including the delete-phase ILM action block named `delete`). The migration SHALL **not** alter the encoding of **`elasticsearch_connection`**.

#### Scenario: Upgrade from list-shaped phase state

- GIVEN persisted state where a phase is stored as a JSON array containing one object
- WHEN Terraform loads state and runs the state upgrader
- THEN the upgraded state SHALL store that phase as a single object (or equivalent null) consistent with `SingleNestedBlock` semantics

#### Scenario: Connection block unchanged by upgrade

- GIVEN persisted state that includes `elasticsearch_connection` as a list
- WHEN the state upgrader runs
- THEN the `elasticsearch_connection` value SHALL remain list-shaped as defined by the connection schema

### Requirement: Action fields optional with object-level AlsoRequires (REQ-022)

For the ILM action blocks **`forcemerge`**, **`searchable_snapshot`**, **`set_priority`**, **`wait_for_snapshot`**, and **`downsample`**, each attribute that is **required for API correctness when the action is declared** SHALL be **optional** at the Terraform attribute level (so an entirely omitted action block does not force those attributes to appear).

When the user **declares** one of these action blocks in configuration, validation SHALL require that all of the following previously required attributes are set (non-null), using object-level validation equivalent to **`objectvalidator.AlsoRequires`**:

- **`forcemerge`**: `max_num_segments`
- **`searchable_snapshot`**: `snapshot_repository`
- **`set_priority`**: `priority`
- **`wait_for_snapshot`**: `policy`
- **`downsample`**: `fixed_interval`

Existing attribute-level validators (for example minimum values) SHALL remain on those attributes where applicable.

#### Scenario: Omitted action block is valid

- GIVEN a phase without a particular action block (for example no `forcemerge` block)
- WHEN Terraform validates configuration
- THEN validation SHALL NOT fail solely because `max_num_segments` is unset

#### Scenario: Empty action block is invalid

- GIVEN the user declares `forcemerge { }` with no attributes
- WHEN Terraform validates configuration
- THEN validation SHALL fail with a diagnostic indicating the required fields when the block is present

#### Scenario: Searchable snapshot requires repository when present

- GIVEN the user declares `searchable_snapshot { force_merge_index = true }` without `snapshot_repository`
- WHEN Terraform validates configuration
- THEN validation SHALL fail with a diagnostic
