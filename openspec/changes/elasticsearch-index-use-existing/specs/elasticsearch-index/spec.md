## ADDED Requirements

### Requirement: Opt-in adoption of existing indices via `use_existing`

The resource SHALL expose an optional boolean attribute `use_existing` (default `false`) on `elasticstack_elasticsearch_index`. The attribute SHALL NOT have any plan modifier that triggers replacement when its value changes; flipping it on an already-managed resource SHALL be a planning no-op. When `use_existing` is `false` or unset, the resource SHALL NOT perform any pre-create existence check and SHALL behave exactly as it does today.

When `use_existing` is `true` and the configured `name` is a static index name (i.e. it does not match the date-math regex), the create flow SHALL, before issuing the Create Index API call, query the Get Index API for that name. If Get Index returns no matching index (HTTP 404 or absent from the response body), the resource SHALL fall through to the normal create path without emitting any extra diagnostic.

If Get Index returns an existing index, the resource SHALL adopt it. Adoption SHALL:

1. Build a synthetic prior state from the Get Index API response using the same response-to-state mapping used by the Read flow.
2. Compare each static index setting attribute that is explicitly set in the plan (i.e. neither null nor unknown) against the value reported by the Get Index API for the same setting key. The static settings compared are those listed in `staticSettingsKeys` (currently: `number_of_shards`, `number_of_routing_shards`, `codec`, `routing_partition_size`, `load_fixed_bitset_filters_eagerly`, `shard.check_on_startup`, `sort.field`, `sort.order`, `mapping.coerce`). Static settings not explicitly set in config SHALL NOT be compared.
3. If any explicitly-configured static setting differs from the existing index, the resource SHALL return an error diagnostic naming every mismatched attribute together with its configured and actual values, SHALL NOT call any Put Settings, Put Mapping, Put Alias, or Delete Alias APIs, and SHALL NOT update Terraform state.
4. Otherwise, reconcile the existing index against the plan by reusing the same alias, dynamic-setting, and mappings reconciliation logic used by the resource's update flow. Aliases present on the existing index but absent from the plan SHALL be deleted via the Delete Alias API; aliases present in the plan SHALL be upserted via the Put Alias API. Dynamic-setting differences SHALL be applied via Put Settings, with dynamic settings present on the existing index but absent from the plan sent as `null`. Mapping differences SHALL be applied via Put Mapping using the resource's existing mapping semantic-equality logic, including the existing template-aware superset handling.
5. Compute `id` from the current cluster UUID and the concrete index name returned by Get Index, set `concrete_name` to that concrete index name, and perform the standard post-create read to refresh state.
6. Emit a warning diagnostic stating that the existing index was adopted, including the concrete index name in the message.

When `use_existing` is `true` and the configured `name` is a plain date math expression (i.e. it matches `DateMathIndexNameRe`), the resource SHALL skip the existence check, emit a warning diagnostic explaining that `use_existing` has no effect for date math names, and proceed along the normal create path without further changes in behavior.

After adoption, the resource SHALL be fully managed by Terraform: subsequent reads, updates, and deletes SHALL follow the existing flows. In particular, a subsequent destroy SHALL call the Delete Index API for the adopted concrete index, gated by `deletion_protection` according to its existing semantics.

#### Scenario: `use_existing = false` keeps the create path unchanged

- **GIVEN** `use_existing` is `false` (or unset) on an `elasticstack_elasticsearch_index` resource
- **WHEN** create runs
- **THEN** the resource SHALL NOT call the Get Index API before the Create Index API
- **AND** the create flow SHALL behave exactly as it does without this attribute

#### Scenario: `use_existing = true` adopts an existing static-named index

- **GIVEN** `use_existing = true`, `name` is a static index name, and an index with that name already exists in Elasticsearch
- **AND** every static setting explicitly set in the plan matches the existing index's static settings
- **WHEN** create runs
- **THEN** the resource SHALL NOT call the Create Index API
- **AND** the resource SHALL reconcile aliases, dynamic settings, and mappings against the existing index using the same logic as a normal update
- **AND** `id` SHALL be set from the cluster UUID and the existing concrete index name
- **AND** `concrete_name` SHALL be set to the existing concrete index name
- **AND** a warning diagnostic indicating that the index was adopted (including the concrete name) SHALL be emitted
- **AND** a subsequent Terraform plan against the same configuration SHALL be empty

#### Scenario: `use_existing = true` falls through to create when the index does not exist

- **GIVEN** `use_existing = true`, `name` is a static index name, and no index with that name exists in Elasticsearch
- **WHEN** create runs
- **THEN** the resource SHALL call the Create Index API as it does today
- **AND** SHALL NOT emit an adoption warning

#### Scenario: Adoption with mismatched static settings fails without mutating the cluster

- **GIVEN** `use_existing = true`, an existing index whose static settings differ from at least one static setting explicitly set in the plan
- **WHEN** create runs
- **THEN** the resource SHALL return an error diagnostic listing every mismatched static setting with its configured and actual values
- **AND** the resource SHALL NOT call Put Settings, Put Mapping, Put Alias, or Delete Alias on the existing index
- **AND** Terraform state SHALL NOT be updated

#### Scenario: Adopting an index reconciles aliases symmetrically with normal updates

- **GIVEN** `use_existing = true` and an existing index that has alias `legacy_alias`
- **AND** the plan defines alias `new_alias` and does not declare `legacy_alias`
- **WHEN** create runs and adoption proceeds
- **THEN** the resource SHALL call the Delete Alias API for `legacy_alias`
- **AND** SHALL call the Put Alias API for `new_alias`

#### Scenario: Adopting an index tolerates template-injected mapping supersets

- **GIVEN** `use_existing = true` and an existing index whose mappings are a non-drifting superset of the plan's mappings due to a matching index template
- **WHEN** create runs and adoption proceeds
- **THEN** the resource SHALL NOT call the Put Mapping API solely for the template-owned differences
- **AND** Terraform SHALL produce an empty plan for the unchanged configuration on the next refresh

#### Scenario: `use_existing = true` with a date math name emits a warning and falls through to create

- **GIVEN** `use_existing = true` and `name` is a plain Elasticsearch date math expression
- **WHEN** create runs
- **THEN** the resource SHALL NOT call the Get Index API for that name
- **AND** the resource SHALL emit a warning diagnostic explaining that `use_existing` has no effect for date math names
- **AND** the resource SHALL proceed with the normal create path (URI-encoded date math name, capture concrete name from response, etc.)

#### Scenario: Flipping `use_existing` after adoption is a planning no-op

- **GIVEN** an `elasticstack_elasticsearch_index` resource has been created via adoption with `use_existing = true`
- **WHEN** the configuration changes `use_existing` from `true` to `false` (or back) without any other changes
- **THEN** Terraform SHALL plan an in-place update with no API calls
- **AND** the resource SHALL NOT be marked for replacement

## MODIFIED Requirements

### Requirement: Create flow (REQ-013–REQ-014)

On create, the resource SHALL build an API model from the plan (including settings, mappings, and aliases) and submit a Create Index request using the configured `name` together with the configured `wait_for_active_shards`, `master_timeout`, and `timeout` parameters. When the configured `name` is a validated date math expression, the provider SHALL URI-encode it before sending the Create Index API request path. After a successful create, the resource SHALL capture the concrete index name returned by the Create Index API response, store it in `concrete_name`, compute `id` from the cluster UUID and that concrete name, and then perform a read to refresh all computed attributes in state. That post-create read SHALL preserve the configured `name` value in state rather than replacing it with the concrete index name.

When `use_existing` is `true` and the configured `name` is a static index name, the resource SHALL first call the Get Index API for that name. If the index already exists, the resource SHALL adopt it as defined by the "Opt-in adoption of existing indices via `use_existing`" requirement, in which case the Create Index API SHALL NOT be called. If the index does not exist, the resource SHALL proceed with the normal create path described above. When `use_existing` is `true` and the configured `name` is a date math expression, the resource SHALL emit a warning diagnostic and proceed with the normal create path without performing any existence check.

#### Scenario: Serverless — master_timeout and wait_for_active_shards omitted

- **GIVEN** the Elasticsearch server flavor is `serverless`
- **WHEN** a create request is issued
- **THEN** `master_timeout` and `wait_for_active_shards` SHALL be omitted from the API call parameters

#### Scenario: Date math create stores configured and concrete names separately

- **WHEN** the configuration uses a plain date math index name and Elasticsearch creates a concrete index from it
- **THEN** state SHALL preserve the configured expression in `name` and store the concrete created index in `concrete_name`

#### Scenario: `use_existing = true` short-circuits the Create Index API for an existing index

- **GIVEN** `use_existing = true`, `name` is a static index name, and the index already exists in Elasticsearch
- **WHEN** create runs
- **THEN** the resource SHALL NOT call the Create Index API
- **AND** the resource SHALL run the adoption flow

#### Scenario: `use_existing = true` falls through to the normal create when the index does not exist

- **GIVEN** `use_existing = true`, `name` is a static index name, and no index with that name exists in Elasticsearch
- **WHEN** create runs
- **THEN** the resource SHALL call the Create Index API as it would when `use_existing = false`
