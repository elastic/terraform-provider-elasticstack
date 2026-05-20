## MODIFIED Schema

### Action block shapes — `allocate`

The `allocate` block schema description in `openspec/specs/elasticsearch-index-lifecycle/spec.md` under "Action block shapes" SHALL be updated to read:

```hcl
allocate {
  number_of_replicas    = <optional + computed, int>    # no default; omitted → not sent to API
  total_shards_per_node = <optional + computed, int>    # no default; omitted → not sent to API
  include               = <optional, json object string>
  exclude               = <optional, json object string>
  require               = <optional, json object string>
}
```

The previous descriptions of `# default 0` and `# default -1` are removed.

## ADDED Requirements

### Requirement: Allocate action omits replica/shard fields when not configured (REQ-034)

The `allocate` action block SHALL treat `number_of_replicas` and `total_shards_per_node` as independently optional with no provider-injected default values.

When either field is absent from the Terraform configuration, the provider SHALL NOT include that field in the Elasticsearch ILM PUT policy request. The Elasticsearch API will leave the current index setting unchanged for that field.

When a field is explicitly set to any value (including `0` for `number_of_replicas` or `-1` for `total_shards_per_node`), the provider SHALL include that field in the API request with the specified value.

Both fields SHALL remain `Computed: true` so that the provider can read back explicit values from the API into Terraform state (for example, when importing an existing policy that has explicit replica counts in its allocate action).

Both fields SHALL use the `UseStateForUnknown` plan modifier so that existing resources whose state already contains `0` or `-1` (previously injected by old provider versions) produce no plan diff on upgrade without explicit changes to the configuration.

#### Scenario: Allocate with routing filter only — no replica/shard fields sent

- GIVEN a Terraform configuration that declares an `allocate` block with only routing-filter attributes (`require`, `include`, and/or `exclude`) and omits `number_of_replicas` and `total_shards_per_node`
- WHEN the provider expands the configuration to the ILM API payload
- THEN the API request body SHALL NOT contain `number_of_replicas` or `total_shards_per_node` keys

#### Scenario: Explicit replica/shard values are preserved

- GIVEN a Terraform configuration that explicitly sets `number_of_replicas = 2` and/or `total_shards_per_node = 100`
- WHEN the provider expands the configuration to the ILM API payload
- THEN the API request body SHALL contain those fields with the specified values

#### Scenario: Explicit zero/negative-one values are preserved

- GIVEN a Terraform configuration that explicitly sets `number_of_replicas = 0` and/or `total_shards_per_node = -1`
- WHEN the provider expands the configuration
- THEN the API request body SHALL contain those fields with the specified values (they are not filtered as empty)

#### Scenario: API response without replica/shard fields is read back cleanly

- GIVEN the Elasticsearch API returns an `allocate` action response that omits `total_shards_per_node`
- WHEN the provider flattens the response into Terraform state
- THEN `total_shards_per_node` in state SHALL be null (not synthetically set to `-1`)

#### Scenario: Upgrade — existing state with injected default values produces no diff

- GIVEN an existing Terraform resource whose state has `number_of_replicas = 0` and `total_shards_per_node = -1` from a previous provider version
- AND the Terraform configuration omits both fields
- WHEN the user runs `terraform plan` after upgrading the provider
- THEN the plan SHALL show no changes for `number_of_replicas` or `total_shards_per_node` (the `UseStateForUnknown` modifier propagates the state values to plan)
