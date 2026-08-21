## MODIFIED Requirements

### Requirement: Version-gated ILM settings (REQ-022–REQ-025)

For ILM action settings that are only supported on newer Elasticsearch versions, the provider SHALL compare the connected server version to the setting's minimum supported version during expansion. If the configured value is non-default on an unsupported server, the provider SHALL return an error diagnostic. If the configured value equals the default, the provider SHALL omit that unsupported setting from the payload instead of failing.

The following minimum versions SHALL apply:

- `rollover.max_primary_shard_docs`: Elasticsearch `8.2.0`
- `rollover.min_age`, `rollover.min_docs`, `rollover.min_size`, `rollover.min_primary_shard_docs`, `rollover.min_primary_shard_size`: Elasticsearch `8.4.0`
- `shrink.allow_write_after_shrink` when `true`: Elasticsearch `8.14.0`
- `searchable_snapshot.force_merge_on_clone` when `false`: Elasticsearch `9.2.1`

ILM settings available throughout the supported `8.x` and later range SHALL NOT have pre-8.0 compatibility gates.

The `searchable_snapshot` action block, available in the `hot`, `cold`, and `frozen` phases, SHALL additionally accept an optional `force_merge_on_clone` boolean attribute (computed default `true`), alongside its existing `snapshot_repository` and `force_merge_index` attributes:

```hcl
searchable_snapshot {
  snapshot_repository  = <optional, string>          # required when block is present
  force_merge_index    = <optional + computed, bool> # default true
  force_merge_on_clone = <optional + computed, bool> # default true; false requires ES >= 9.2.1
}
```

`force_merge_on_clone` SHALL be rejected by config validation when configured (to any value) together with `force_merge_index = false`, since Elasticsearch disallows a non-null `force_merge_on_clone` in that combination. When `force_merge_index` is `false` and `force_merge_on_clone` is left unset, the provider SHALL omit `force_merge_on_clone` from the API payload regardless of its computed default.

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

#### Scenario: Unsupported searchable-snapshot force-merge-on-clone override

- GIVEN a `searchable_snapshot` block configures `force_merge_on_clone = false`
- AND the connected Elasticsearch server is below `9.2.1`
- WHEN the policy is expanded
- THEN the provider SHALL return an unsupported-setting diagnostic
- AND the provider SHALL NOT call the Put Lifecycle API

#### Scenario: Default searchable-snapshot force-merge-on-clone omitted pre-9.2.1

- GIVEN a `searchable_snapshot` block leaves `force_merge_on_clone` unset (its default `true`)
- AND the connected Elasticsearch server is below `9.2.1`
- WHEN the policy is expanded
- THEN the provider SHALL omit `force_merge_on_clone` from the API payload
- AND the Put Lifecycle API call SHALL proceed

#### Scenario: Searchable-snapshot force-merge-on-clone sent on supported servers

- GIVEN a `searchable_snapshot` block configures `force_merge_on_clone = false`
- AND the connected Elasticsearch server is `9.2.1` or above
- WHEN the policy is expanded
- THEN the provider SHALL include `force_merge_on_clone: false` in the `searchable_snapshot` action payload

#### Scenario: Force-merge-on-clone rejected when force-merge-index is disabled

- GIVEN a `searchable_snapshot` block configures `force_merge_index = false`
- AND the same block also configures `force_merge_on_clone` (to any value)
- WHEN the configuration is validated
- THEN the provider SHALL return a diagnostic rejecting `force_merge_on_clone` alongside `force_merge_index = false`
- AND the provider SHALL NOT call the Put Lifecycle API

#### Scenario: Force-merge-on-clone omitted when force-merge-index is disabled and left unset

- GIVEN a `searchable_snapshot` block configures `force_merge_index = false`
- AND the same block leaves `force_merge_on_clone` unset
- WHEN the policy is expanded against a supported (`9.2.1`+) Elasticsearch server
- THEN the provider SHALL omit `force_merge_on_clone` from the `searchable_snapshot` action payload
- AND the Put Lifecycle API call SHALL proceed

#### Scenario: Force-merge-on-clone default backfilled on read when omitted by Elasticsearch

- GIVEN a `searchable_snapshot` block has `force_merge_index = true` and `force_merge_on_clone` left unset (its default `true`)
- AND the Get Lifecycle API response omits `force_merge_on_clone` from the `searchable_snapshot` action, as Elasticsearch does whenever the field was never explicitly set
- WHEN the provider reads the policy into state
- THEN the provider SHALL populate `force_merge_on_clone` as `true` in state
- AND SHALL NOT leave it null

### Requirement: Create and update flow (REQ-011–REQ-012)

Create and update SHALL expand the Terraform model into a full `models.Policy`, set `policy.Name` from `name`, submit the policy with the Put Lifecycle API, set `id`, and then read the policy back so computed fields and cluster-returned values are refreshed into state.

The Put Lifecycle API call SHALL be made by marshaling the policy to JSON and submitting that JSON as the request's raw body, rather than by unmarshaling it into the Elasticsearch client library's typed lifecycle-put request type and submitting that typed value. This is required because the typed request type does not model every action-level field the provider's schema supports (for example `searchable_snapshot.force_merge_on_clone`); submitting a typed value would silently drop such fields before they reach Elasticsearch.

The subsequent Get Lifecycle read (both the read-after-write performed by create/update, and any later read/refresh) SHALL likewise avoid decoding the response through the Elasticsearch client library's typed lifecycle response type for action fields the typed type does not model. This is required for the same reason as the write-path change above, and symmetrically: the typed response type's action structs (for example `searchable_snapshot`'s) silently discard unrecognized JSON keys during decoding, so a field successfully written via the raw-body path would otherwise be silently dropped again on the very next read, leaving Terraform state showing the field's default instead of the value actually stored in Elasticsearch.

#### Scenario: Read after successful put

- GIVEN a successful Put Lifecycle request
- WHEN create or update completes
- THEN the provider SHALL perform read-after-write and populate computed state such as `modified_date`

#### Scenario: Action field unsupported by the typed client library still reaches Elasticsearch

- GIVEN a phase action configures a field that the Elasticsearch client library's typed lifecycle-put request type does not declare (for example `searchable_snapshot.force_merge_on_clone`)
- AND the field is supported by the connected Elasticsearch server version
- WHEN create or update expands and submits the policy
- THEN the field SHALL be present in the raw JSON body submitted to the Put Lifecycle API
- AND the field SHALL NOT be silently dropped by typed (de)serialization before submission

#### Scenario: Action field unsupported by the typed client library survives read-after-write

- GIVEN a phase action configures a field that the Elasticsearch client library's typed lifecycle-get response type does not declare (for example `searchable_snapshot.force_merge_on_clone`)
- AND the field is supported by the connected Elasticsearch server version
- AND create or update has successfully submitted a policy containing that field with a non-default value
- WHEN the provider performs the read-after-write (or any subsequent read/refresh)
- THEN the field's configured value SHALL be present in the resulting Terraform state
- AND the field SHALL NOT be silently reverted to its default by typed (de)serialization of the Get Lifecycle response
