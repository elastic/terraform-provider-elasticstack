# `elasticstack_elasticsearch_index_lifecycle` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/ilm/`

## Purpose

Manage Elasticsearch Index Lifecycle Management (ILM) policies through the Terraform Plugin Framework resource. The resource creates and updates policies, reads them back into Terraform state, deletes them by policy name, supports import by composite id, allows an optional resource-scoped Elasticsearch connection, and preserves compatibility with older Elasticsearch versions and older SDK-shaped state.

## Schema

### Top-level attributes

```hcl
resource "elasticstack_elasticsearch_index_lifecycle" "example" {
  id            = <computed, string>      # <cluster_uuid>/<policy_name>
  name          = <required, string>      # force new
  metadata      = <optional, json object> # normalized JSON string
  force_destroy = <optional + computed, bool> # default false
  modified_date = <computed, string>

  hot    { /* SingleNestedBlock */ }
  warm   { /* SingleNestedBlock */ }
  cold   { /* SingleNestedBlock */ }
  frozen { /* SingleNestedBlock */ }
  delete { /* SingleNestedBlock */ }

  elasticsearch_connection { /* list nested block from shared provider schema */ }
}
```

### Phase blocks

Each phase block is a Plugin Framework `SingleNestedBlock`. In Terraform state, each phase is stored as a single object value or `null`, not as a single-element list.

Every phase supports:

| Attribute | Type | Notes |
|-----------|------|-------|
| `min_age` | optional + computed string | Defaults to `"0ms"` when omitted. |

Allowed nested action blocks:

| Phase | Allowed actions |
|-------|-----------------|
| `hot` | `set_priority`, `unfollow`, `rollover`, `readonly`, `shrink`, `forcemerge`, `searchable_snapshot`, `downsample` |
| `warm` | `set_priority`, `unfollow`, `readonly`, `allocate`, `migrate`, `shrink`, `forcemerge`, `downsample` |
| `cold` | `set_priority`, `unfollow`, `readonly`, `searchable_snapshot`, `allocate`, `migrate`, `freeze`, `downsample` |
| `frozen` | `searchable_snapshot` (required when `frozen` is declared) |
| `delete` | `wait_for_snapshot`, `delete` |

### Action block shapes

All ILM action blocks are also `SingleNestedBlock`s.

```hcl
allocate {
  number_of_replicas    = <optional + computed, int>    # default 0
  total_shards_per_node = <optional + computed, int>    # default -1
  include               = <optional, json object string>
  exclude               = <optional, json object string>
  require               = <optional, json object string>
}

delete {
  delete_searchable_snapshot = <optional + computed, bool> # default true
}

forcemerge {
  max_num_segments = <optional, int>    # required when block is present; >= 1
  index_codec      = <optional, string>
}

freeze {
  enabled = <optional + computed, bool> # default true
}

migrate {
  enabled = <optional + computed, bool> # default true
}

readonly {
  enabled = <optional + computed, bool> # default true
}

rollover {
  max_age                = <optional, string>
  max_docs               = <optional, int>
  max_size               = <optional, string>
  max_primary_shard_docs = <optional, int>    # non-default requires ES >= 8.2
  max_primary_shard_size = <optional, string>
  min_age                = <optional, string> # non-default requires ES >= 8.4
  min_docs               = <optional, int>    # non-default requires ES >= 8.4
  min_size               = <optional, string> # non-default requires ES >= 8.4
  min_primary_shard_docs = <optional, int>    # non-default requires ES >= 8.4
  min_primary_shard_size = <optional, string> # non-default requires ES >= 8.4
}

searchable_snapshot {
  snapshot_repository = <optional, string>          # required when block is present
  force_merge_index   = <optional + computed, bool> # default true
}

set_priority {
  priority = <optional, int> # required when block is present; >= 0
}

shrink {
  number_of_shards         = <optional, int>
  max_primary_shard_size   = <optional, string>
  allow_write_after_shrink = <optional + computed, bool> # default false; non-default requires ES >= 8.14
}

unfollow {
  enabled = <optional + computed, bool> # default true
}

wait_for_snapshot {
  policy = <optional, string> # required when block is present
}

downsample {
  fixed_interval = <optional, string> # required when block is present
  wait_timeout   = <optional + computed, string>
}
```

Additional schema behavior:

- When the `frozen` phase is declared, the `searchable_snapshot` nested block is required in the Terraform schema (unlike `hot` and `cold`, where that action remains optional).
- `metadata`, `allocate.include`, `allocate.exclude`, and `allocate.require` use normalized JSON object string types and validate JSON-object syntax.
- Empty allocation filter objects are omitted from state on read so unset optional filters remain absent.
- `elasticsearch_connection` remains list-shaped in state because it comes from the shared provider connection schema.
## Requirements
### Requirement: CRUD APIs and diagnostics (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Lifecycle API to create and update ILM policies, the Get Lifecycle API to read them, and the Delete Lifecycle API to delete them. When Elasticsearch returns a non-success response for create, update, read, or delete, except for HTTP `404` on read, the resource SHALL surface that failure as Terraform diagnostics.

#### Scenario: Non-success lifecycle API response

- GIVEN Elasticsearch returns a non-success response for Put, Get, or Delete lifecycle
- WHEN the provider handles that response
- THEN Terraform SHALL receive an error diagnostic

### Requirement: Identity, import, and replacement (REQ-005–REQ-007)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`. Create and update SHALL derive that id from the connected cluster UUID and the configured `name`. Import SHALL use passthrough of the provided `id`. Changing `name` SHALL require replacement instead of in-place rename.

#### Scenario: Import by composite id

- GIVEN an import id in the form `<cluster_uuid>/<policy_name>`
- WHEN import completes
- THEN the resource SHALL store that id unchanged and use it on subsequent read and delete

#### Scenario: Rename requested

- GIVEN an existing resource instance
- WHEN `name` changes in configuration
- THEN Terraform SHALL plan replacement

### Requirement: Validation and connection selection (REQ-008–REQ-010)

The resource SHALL reject configuration that omits all five phase blocks `hot`, `warm`, `cold`, `frozen`, and `delete`. The resource SHALL accept `metadata` and allocation filters only when they are valid JSON objects. By default, the resource SHALL use the provider-level Elasticsearch client; when `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for create, read, update, and delete.

When the user declares the `frozen` phase, the configuration SHALL include a `searchable_snapshot` block inside `frozen`; omission SHALL be rejected during Terraform validation before any lifecycle API call.

#### Scenario: No lifecycle phases configured

- GIVEN all phase blocks are absent
- WHEN configuration is validated
- THEN the provider SHALL return a validation error before any lifecycle API call

#### Scenario: Frozen phase without searchable snapshot is rejected

- GIVEN a resource configuration with `frozen { min_age = "30d" }` and no `searchable_snapshot`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error before any Elasticsearch ILM API call

#### Scenario: Resource-scoped connection override

- GIVEN `elasticsearch_connection` is configured for the resource
- WHEN CRUD operations run
- THEN the provider SHALL use that resource-scoped connection

### Requirement: Create and update flow (REQ-011–REQ-012)

Create and update SHALL expand the Terraform model into a full `models.Policy`, set `policy.Name` from `name`, submit the policy with the Put Lifecycle API, set `id`, and then read the policy back so computed fields and cluster-returned values are refreshed into state.

#### Scenario: Read after successful put

- GIVEN a successful Put Lifecycle request
- WHEN create or update completes
- THEN the provider SHALL perform read-after-write and populate computed state such as `modified_date`

### Requirement: Read and delete behavior (REQ-013–REQ-016)

Read SHALL interpret `id` as the composite identifier `<cluster_uuid>/<policy_name>`. If `id` is not in that format or does not contain a policy name portion, Read SHALL return a Terraform diagnostic describing the invalid id and SHALL NOT issue Elasticsearch API requests.

For a valid composite `id`, Read SHALL call the Get Lifecycle API using the policy name portion of `id`.

If the Get Lifecycle API returns a not found response for that policy, Read SHALL remove the resource from Terraform state.

If the Get Lifecycle API returns success but the response body does not contain an entry for the requested policy name, Read SHALL return a Terraform diagnostic describing the missing policy in the response and SHALL NOT write partial state.

Otherwise, Read SHALL map the returned policy into Terraform state, preserving computed fields and using the existing composite `id`.

#### Scenario: Invalid composite id during read

- GIVEN the resource `id` is not a valid `<cluster_uuid>/<policy_name>` composite id
- WHEN read runs
- THEN the provider SHALL return a diagnostic about the invalid id
- AND the provider SHALL NOT call the Get Lifecycle API

#### Scenario: ILM policy not found during read

- GIVEN the resource `id` contains a valid policy name
- AND the Get Lifecycle API returns not found for that policy
- WHEN read runs
- THEN the provider SHALL remove the resource from Terraform state

#### Scenario: Requested policy missing from Get Lifecycle response

- GIVEN the resource `id` contains a valid policy name
- AND the Get Lifecycle API returns success
- AND the response does not include the requested policy name
- WHEN read runs
- THEN the provider SHALL return a diagnostic about the missing policy in the response
- AND the provider SHALL NOT write partial state

Delete SHALL call the Delete Lifecycle API with the policy name portion of `id`.

When `force_destroy` is `true`, Delete SHALL additionally identify any indices whose `index.lifecycle.name` setting references the policy name and remove that reference by setting `index.lifecycle.name` to `null` before invoking the Delete Lifecycle API. The scan-and-clear process SHALL be:

1. Query `GET /_all/_settings/index.lifecycle.name?flat_settings=true` to obtain the `index.lifecycle.name` setting for every index.
2. Filter to indices whose setting value equals the policy name being deleted.
3. If one or more indices match, issue `PUT /{indices}/_settings` with `{"index.lifecycle.name": null}` where `{indices}` is a comma-separated list of matched index names.
4. After clearing references, proceed with `DELETE /_ilm/policy/{policy_name}`.

If the settings-clear call returns an error, Delete SHALL surface that error as a Terraform diagnostic and SHALL NOT proceed with the Delete Lifecycle API call. If the subsequent Delete Lifecycle API call returns an error (for example, because a new index referencing the policy was created during the clear step), Delete SHALL surface the Elasticsearch error verbatim.

When `force_destroy` is `false` (the default), Delete SHALL call the Delete Lifecycle API directly without scanning indices. If Elasticsearch rejects the delete because the policy is still in use, the provider SHALL surface that error verbatim.

#### Scenario: ILM policy deleted while referenced by backing index (force_destroy = true)

- GIVEN an ILM policy named `"my-policy"` exists with `force_destroy = true`
- AND an index `".ds-logs-test-default-2026.01.01-000001"` has `index.lifecycle.name` set to `"my-policy"`
- WHEN Delete runs for the ILM policy resource
- THEN the provider SHALL first set `index.lifecycle.name` to `null` on `.ds-logs-test-default-2026.01.01-000001`
- AND then call `DELETE /_ilm/policy/my-policy`
- AND the resource SHALL be destroyed successfully

#### Scenario: No indices reference the policy (force_destroy = true)

- GIVEN an ILM policy named `"unused-policy"` exists with `force_destroy = true`
- AND no index has `index.lifecycle.name` set to `"unused-policy"`
- WHEN Delete runs for the ILM policy resource
- THEN the provider SHALL skip the settings-clear step
- AND call `DELETE /_ilm/policy/unused-policy` directly

#### Scenario: Settings-clear fails before delete (force_destroy = true)

- GIVEN an ILM policy named `"my-policy"` exists with `force_destroy = true`
- AND an index referencing the policy exists
- AND the `PUT /_settings` call to clear the reference fails (e.g., index is closed or unavailable)
- WHEN Delete runs
- THEN the provider SHALL surface the settings-clear error as a Terraform diagnostic
- AND SHALL NOT call `DELETE /_ilm/policy/my-policy`

#### Scenario: ILM policy delete fails when referenced and force_destroy = false

- GIVEN an ILM policy named `"my-policy"` exists with `force_destroy = false`
- AND an index `"test-index"` has `index.lifecycle.name` set to `"my-policy"`
- WHEN Delete runs for the ILM policy resource
- THEN the provider SHALL call `DELETE /_ilm/policy/my-policy` directly
- AND Elasticsearch SHALL reject the request because the policy is in use
- AND the provider SHALL surface the Elasticsearch error verbatim

### Requirement: Metadata and phase/action mapping (REQ-017–REQ-021)

On create and update, the resource SHALL decode `metadata` JSON into the policy metadata map when `metadata` is set. For each configured phase, the resource SHALL expand `min_age` and the configured action blocks into the API model. `allocate.include`, `allocate.exclude`, and `allocate.require` SHALL be decoded from JSON object strings into maps. `readonly`, `freeze`, and `unfollow` SHALL be omitted from the API payload when `enabled = false`. `migrate` SHALL still be sent with its configured `enabled` value, including `false`. Unknown action names encountered during expansion SHALL return an error diagnostic.

On read, the provider SHALL flatten API phases back into Terraform phase objects, serialize allocation filters back to JSON strings, retain prior `metadata` when the API omits metadata, and set `modified_date` from the policy definition returned by Elasticsearch.

#### Scenario: Disabled readonly omitted from API payload

- GIVEN `readonly { enabled = false }` is configured in a phase
- WHEN the policy is expanded for create or update
- THEN the `readonly` action SHALL be omitted from the Elasticsearch payload

#### Scenario: Migrate false preserved in payload

- GIVEN `migrate { enabled = false }` is configured
- WHEN the policy is expanded for create or update
- THEN the payload SHALL still contain the `migrate` action with `enabled = false`

### Requirement: Version-gated ILM settings (REQ-022–REQ-025)

For ILM action settings that are only supported on newer Elasticsearch versions, the provider SHALL compare the connected server version to the setting's minimum supported version during expansion. If the configured value is non-default on an unsupported server, the provider SHALL return an error diagnostic. If the configured value equals the default, the provider SHALL omit that unsupported setting from the payload instead of failing.

The following minimum versions SHALL apply:

- `rollover.max_primary_shard_docs`: Elasticsearch `8.2.0`
- `rollover.min_age`, `rollover.min_docs`, `rollover.min_size`, `rollover.min_primary_shard_docs`, `rollover.min_primary_shard_size`: Elasticsearch `8.4.0`
- `shrink.allow_write_after_shrink` when `true`: Elasticsearch `8.14.0`

ILM settings available throughout the supported `8.x` and later range SHALL NOT have pre-8.0 compatibility gates.

#### Scenario: Unsupported rollover min condition

- GIVEN `rollover.min_docs` is configured with a non-default value
- AND the connected Elasticsearch server is below `8.4.0`
- WHEN the policy is expanded
- THEN the provider SHALL return an unsupported-setting diagnostic

#### Scenario: Unsupported shrink allow-write-after-shrink

- GIVEN the connected Elasticsearch server is below `8.14.0`
- WHEN `shrink.allow_write_after_shrink = true` is configured
- THEN the provider SHALL return an unsupported-setting diagnostic

#### Scenario: Supported-range allocate setting is sent

- GIVEN `allocate.total_shards_per_node` is configured with a value other than `-1`
- WHEN the policy is expanded against a supported Elasticsearch server version
- THEN the provider SHALL include `total_shards_per_node` in the API payload

### Requirement: Read-state normalization (REQ-026–REQ-028)

On read, when the API omits `total_shards_per_node` inside an `allocate` action, the provider SHALL store `-1` in state. When a `shrink` action is present and the API omits `allow_write_after_shrink`, the provider SHALL store `false` in state. When allocation filters serialize to empty JSON objects, the provider SHALL omit those filter attributes from state so unset optional filters remain absent.

#### Scenario: Allocate default restored on read

- GIVEN an `allocate` action from Elasticsearch that omits `total_shards_per_node`
- WHEN the provider flattens the phase
- THEN state SHALL contain `total_shards_per_node = -1`

#### Scenario: Empty allocation filter omitted

- GIVEN an `allocate` action whose `include`, `exclude`, or `require` values serialize to `{}`
- WHEN the provider flattens the phase
- THEN the corresponding Terraform attribute SHALL be absent from state

### Requirement: Disabled toggle preservation across refresh (REQ-029)

For `readonly`, `freeze`, and `unfollow`, when the API omits the action because it is inactive but the prior Terraform configuration had declared the block, the provider SHALL preserve that declaration in state by writing the block with `enabled = false`.

#### Scenario: Disabled unfollow remains disabled after refresh

- GIVEN prior Terraform state declared `unfollow { enabled = false }`
- WHEN refresh reads a phase whose API actions omit `unfollow`
- THEN state SHALL still contain `unfollow.enabled = false`

### Requirement: Plugin Framework nested-block shape and state upgrade (REQ-030–REQ-031)

The resource SHALL model each phase block and each ILM action block as a Plugin Framework `SingleNestedBlock`, so state stores them as objects instead of singleton lists. The resource SHALL use schema version `1` and implement state upgrade from schema version `0`, unwrapping legacy singleton-list phase values and legacy singleton-list action values into object values. The upgrade SHALL leave `elasticsearch_connection` list-shaped.

#### Scenario: Upgrade old SDK-shaped nested values

- GIVEN persisted schema version `0` state with a phase stored as `[ { ... } ]`
- WHEN Terraform runs the state upgrader
- THEN the upgraded state SHALL store that phase as a single object value

#### Scenario: Connection block not rewritten

- GIVEN persisted state with `elasticsearch_connection` stored as a list
- WHEN the ILM state upgrader runs
- THEN `elasticsearch_connection` SHALL remain list-shaped

### Requirement: Action block presence validation (REQ-032)

The blocks `forcemerge`, `searchable_snapshot`, `set_priority`, `wait_for_snapshot`, and `downsample` SHALL keep their key attributes optional at the attribute level so omitted blocks are valid, but SHALL require those attributes when the block is present using object-level validation equivalent to `objectvalidator.AlsoRequires`.

The required-when-present attributes SHALL be:

- `forcemerge.max_num_segments`
- `searchable_snapshot.snapshot_repository`
- `set_priority.priority`
- `wait_for_snapshot.policy`
- `downsample.fixed_interval`

#### Scenario: Empty searchable snapshot block

- GIVEN the user declares `searchable_snapshot { force_merge_index = true }`
- WHEN Terraform validates the block
- THEN validation SHALL fail because `snapshot_repository` is required when the block is present

### Requirement: Frozen phase requires searchable snapshot (REQ-033)

When the `frozen` phase is configured, the resource SHALL require the `frozen.searchable_snapshot` nested block in the Terraform schema rather than treating it as optional.

Within that required block, `snapshot_repository` SHALL remain required when the `searchable_snapshot` block is present, consistent with REQ-032.

The generated Terraform documentation for the resource SHALL reflect this schema shape by describing `frozen.searchable_snapshot` as required within the `frozen` phase.

#### Scenario: Valid frozen phase includes searchable snapshot

- GIVEN a resource configuration with:
  - `frozen.min_age = "30d"`
  - `frozen.searchable_snapshot.snapshot_repository = "repo-a"`
- WHEN Terraform plans or applies the resource
- THEN the provider SHALL accept the `frozen` phase schema shape
- AND the lifecycle policy expansion SHALL include the `searchable_snapshot` action for the `frozen` phase

#### Scenario: Required nested field within frozen searchable snapshot

- GIVEN a resource configuration with `frozen.searchable_snapshot { force_merge_index = false }`
- WHEN Terraform validates the configuration
- THEN validation SHALL fail because `snapshot_repository` is required when the `searchable_snapshot` block is present

#### Scenario: Generated docs match frozen schema requirement

- GIVEN the provider documentation is generated from the resource schema
- WHEN the `elasticstack_elasticsearch_index_lifecycle` docs are refreshed
- THEN the `frozen` section SHALL describe `searchable_snapshot` as required within `frozen`

