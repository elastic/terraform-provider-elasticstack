## MODIFIED Requirements

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
