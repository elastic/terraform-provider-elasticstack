# `elasticstack_elasticsearch_index` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/index`

## Purpose

Define schema and behavior for the Elasticsearch index resource: API usage, identity/import, connection, lifecycle (static vs dynamic settings, deletion protection), mappings plan modifier, alias management, and settings mapping between Terraform state and the Elasticsearch indices API.

## Schema

```hcl
resource "elasticstack_elasticsearch_index" "example" {
  id            = <computed, string> # internal identifier: <cluster_uuid>/<concrete_index_name>
  name          = <required, string> # force new; 1–255 chars; either static lowercase alphanumeric + selected punctuation (cannot start with -, _, +), or a plain date math expression enclosed in angle brackets with at least one {…} section
  concrete_name = <computed, string> # UseStateForUnknown; the concrete Elasticsearch index managed by this resource; equals name for static names; resolved concrete index for date math names

  # Aliases
  alias = <optional+computed, set(object)> { # UseStateForUnknown
    name           = <required, string>
    filter         = <optional, json string (normalized)>
    index_routing  = <optional+computed, string> # UseStateForUnknown; default ""
    is_hidden      = <optional+computed, bool>   # UseStateForUnknown; default false
    is_write_index = <optional+computed, bool>   # UseStateForUnknown; default false
    routing        = <optional+computed, string> # UseStateForUnknown; default ""
    search_routing = <optional+computed, string> # UseStateForUnknown; default ""
  }

  # Static settings (force new on change)
  number_of_shards                = <optional, int64> # force new
  number_of_routing_shards        = <optional, int64> # force new
  codec                           = <optional, string> # force new; allowed: "best_compression"
  routing_partition_size          = <optional, int64>  # force new
  load_fixed_bitset_filters_eagerly = <optional, bool> # force new
  shard_check_on_startup          = <optional, string> # force new; allowed: "false", "true", "checksum"
  sort_field                      = <optional, set(string)> # force new
  sort_order                      = <optional, list(string)> # force new
  mapping_coerce                  = <optional, bool> # force new

  # Dynamic settings (updatable in place)
  number_of_replicas              = <optional, int64>
  auto_expand_replicas            = <optional, string>
  search_idle_after               = <optional, string>
  refresh_interval                = <optional, string>
  max_result_window               = <optional, int64>
  max_inner_result_window         = <optional, int64>
  max_rescore_window              = <optional, int64>
  max_docvalue_fields_search      = <optional, int64>
  max_script_fields               = <optional, int64>
  max_ngram_diff                  = <optional, int64>
  max_shingle_diff                = <optional, int64>
  max_refresh_listeners           = <optional, int64>
  analyze_max_token_count         = <optional, int64>
  highlight_max_analyzed_offset   = <optional, int64>
  max_terms_count                 = <optional, int64>
  max_regex_length                = <optional, int64>
  mapping_total_fields_limit      = <optional, int64>
  query_default_field             = <optional, set(string)>
  routing_allocation_enable       = <optional, string> # allowed: "all", "primaries", "new_primaries", "none"
  routing_rebalance_enable        = <optional, string> # allowed: "all", "primaries", "replicas", "none"
  gc_deletes                      = <optional, string>
  blocks_read_only                = <optional, bool>
  blocks_read_only_allow_delete   = <optional, bool>
  blocks_read                     = <optional, bool>
  blocks_write                    = <optional, bool>
  blocks_metadata                 = <optional, bool>
  default_pipeline                = <optional, string>
  final_pipeline                  = <optional, string>
  unassigned_node_left_delayed_timeout = <optional, string>
  search_slowlog_threshold_query_warn  = <optional, string>
  search_slowlog_threshold_query_info  = <optional, string>
  search_slowlog_threshold_query_debug = <optional, string>
  search_slowlog_threshold_query_trace = <optional, string>
  search_slowlog_threshold_fetch_warn  = <optional, string>
  search_slowlog_threshold_fetch_info  = <optional, string>
  search_slowlog_threshold_fetch_debug = <optional, string>
  search_slowlog_threshold_fetch_trace = <optional, string>
  search_slowlog_level                 = <optional, string> # allowed: "warn", "info", "debug", "trace"
  indexing_slowlog_threshold_index_warn  = <optional, string>
  indexing_slowlog_threshold_index_info  = <optional, string>
  indexing_slowlog_threshold_index_debug = <optional, string>
  indexing_slowlog_threshold_index_trace = <optional, string>
  indexing_slowlog_level                 = <optional, string> # allowed: "warn", "info", "debug", "trace"
  indexing_slowlog_source                = <optional, string>

  # Analysis settings (applied only at create time; not sent on update)
  analysis_analyzer    = <optional, json object string (normalized)>
  analysis_tokenizer   = <optional, json object string (normalized)>
  analysis_char_filter = <optional, json object string (normalized)>
  analysis_filter      = <optional, json object string (normalized)>
  analysis_normalizer  = <optional, json object string (normalized)>

  # Mappings
  mappings = <optional+computed, json object string (normalized)> # UseStateForUnknown + mappings plan modifier

  # Computed-only
  settings_raw = <computed, json string (normalized)> # all raw settings from cluster

  # Operational
  deletion_protection  = <optional+computed, bool>   # default true; prevents destroy
  wait_for_active_shards = <optional+computed, string> # default "1"
  master_timeout       = <optional+computed, duration string> # default "30s"
  timeout              = <optional+computed, duration string> # default "30s"

  # Deprecated
  settings { # deprecated block; max 1; at least 1 setting required
    setting { # set; at least 1 required
      name  = <required, string>
      value = <required, string>
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
### Requirement: Index name validation for static and date math names

The `name` attribute on `elasticstack_elasticsearch_index` SHALL accept either a static index name that matches the existing lowercase index-name rules or a plain Elasticsearch date math index name expression. Validation SHALL keep these paths separate by using `stringvalidator.Any(...)` with the static-name regex `^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$` and the date-math regex `^<[^-_+][a-z0-9!$%&'()+.;=@[\]^{}~_-]*\{[^<>]+\}>$`. The date-math regex enforces that: the name cannot start with `-`, `_`, or `+`; the static prefix consists only of valid index-name characters; and the date math section `{...}` appears at the end immediately before the closing `>` (no suffix after the brace). Values that satisfy neither regex branch SHALL be rejected during schema validation. When a validated date math name is used to create an index, the provider SHALL URI-encode that name before sending it in the Create Index API path.

#### Scenario: Static index names remain valid
- **WHEN** the configuration supplies a static index name that satisfies the existing lowercase-name rules
- **THEN** schema validation SHALL accept the value without requiring date math syntax

#### Scenario: Plain date math index names are accepted
- **WHEN** the configuration supplies a plain date math expression for the index path
- **THEN** schema validation SHALL accept the value without weakening the static-name validator

#### Scenario: Invalid date math syntax is rejected
- **WHEN** the configuration supplies a value that does not satisfy the static-name validator and is not valid for the dedicated date-math validator
- **THEN** schema validation SHALL reject the value before any API call is made

#### Scenario: Provider encodes date math name for create request
- **WHEN** the configuration supplies a valid plain date math name and the provider constructs the Create Index API request
- **THEN** the provider SHALL URI-encode that name in the request path sent to Elasticsearch

### Requirement: Index CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Create Index API to create an index ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html)). The resource SHALL use the Elasticsearch Get Index API to read an index ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-index.html)). On update, the resource SHALL use dedicated APIs: the Put Settings API for dynamic setting changes ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-update-settings.html)), the Put Mapping API for mapping changes ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-mapping.html)), and the Put Alias / Delete Alias APIs for alias changes ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html)). The resource SHALL use the Delete Index API to delete an index ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-index.html)). When Elasticsearch returns a non-success status for any create, update, read, or delete request (other than 404 on read), the resource SHALL surface the API error as a Terraform diagnostic.

#### Scenario: API failure on create

- GIVEN the Create Index API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error and the resource SHALL not be stored in state

#### Scenario: API failure on update

- GIVEN any update API (Put Settings, Put Mapping, Put Alias, Delete Alias) returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<concrete_index_name>` and a computed `concrete_name` attribute containing the concrete Elasticsearch index managed by the resource. During create, the resource SHALL compute `id` from the current cluster UUID and the concrete index name returned by Elasticsearch, not from the configured `name`. For imported or legacy state that lacks `concrete_name`, the resource SHALL derive `concrete_name` from `id.ResourceID` during read and store it in state.

#### Scenario: Id and concrete_name set on create

- **WHEN** a Create Index API call succeeds and Elasticsearch returns the created index name
- **THEN** `concrete_name` SHALL be set to that concrete index name and `id` SHALL be set to `<cluster_uuid>/<concrete_index_name>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import by accepting an `id` value directly via `ImportStatePassthroughID`, persisting the imported `id` to state without validation at import time. Read and delete operations SHALL parse `id` in the format `<cluster_uuid>/<concrete_index_name>` and SHALL return an error diagnostic when the format is invalid. When imported or legacy state lacks `concrete_name`, read SHALL backfill it from `id.ResourceID`. When imported state also lacks `name`, read SHALL backfill `name` from the concrete index name so the resource remains readable without inventing a date math expression.

#### Scenario: Import passthrough backfills concrete identity

- **WHEN** an import command stores a composite `id` and the next read runs
- **THEN** the resource SHALL use the imported resource id as the concrete index identity for subsequent read, update, and delete operations

### Requirement: Lifecycle — static settings require replacement (REQ-009)

In addition to the existing static settings listed in REQ-009, entries in the new `sort` `ListNestedAttribute` SHALL also trigger resource replacement when changed, subject to the migration suppression defined in REQ-SORT-03. Specifically, changing any of `sort[*].field`, `sort[*].order`, `sort[*].missing`, or `sort[*].mode` SHALL require replacement. The deprecated `sort_field` attribute SHALL continue to require replace when changed, except when `sort` is simultaneously being introduced in the plan (REQ-SORT-03 governs in that case).

#### Scenario: Changing an existing `sort` entry's `order` requires replace

- **GIVEN** an existing index managed with `sort = [{ field = "date", order = "asc" }]`
- **WHEN** the configuration changes to `sort = [{ field = "date", order = "desc" }]`
- **THEN** Terraform SHALL plan to destroy and recreate the resource

---

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When the `elasticsearch_connection` block is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set with valid credentials
- WHEN the resource performs any API call
- THEN it SHALL use the resource-scoped client, not the provider-level client

### Requirement: Deletion protection (REQ-012)

When `deletion_protection` is `true` (the default), the resource SHALL refuse to delete the index and SHALL return an error diagnostic on the `deletion_protection` attribute instructing the user to set it to `false` and apply before destroying. When `deletion_protection` is `false`, the resource SHALL proceed with deletion.

#### Scenario: Deletion blocked

- GIVEN `deletion_protection = true` (default)
- WHEN a destroy or apply that removes the resource is run
- THEN the resource SHALL fail with an error diagnostic on `deletion_protection` and SHALL not call the Delete Index API

#### Scenario: Deletion allowed

- GIVEN `deletion_protection = false`
- WHEN a destroy or apply that removes the resource is run
- THEN the resource SHALL call the Delete Index API and remove the resource from state

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

### Requirement: Update flow (REQ-015–REQ-018)

On update, the resource SHALL only call the relevant update APIs when the corresponding values have changed. Alias changes SHALL be applied by deleting aliases removed from config (via Delete Alias API) and upserting all aliases present in plan (via Put Alias API). Dynamic setting changes SHALL be applied by calling the Put Settings API with the diff, setting removed dynamic settings to `null` in the request. Mapping changes SHALL be applied by calling the Put Mapping API only when the user-owned mapping intent has semantically changed. Template-injected mapping content that appears in the Elasticsearch Get Index API response SHALL NOT by itself cause a mapping update, replacement, provider inconsistent-result error, or non-empty follow-up plan. All update APIs SHALL target the persisted concrete index identity from state / `id`, not the configured `name`. After all updates, the resource SHALL perform a read to refresh state while preserving any configured `name` already stored in state.

#### Scenario: Removed alias is deleted

- WHEN state has alias `old_alias` and config does not
- THEN update SHALL call the Delete Alias API for `old_alias`

#### Scenario: Removed dynamic setting is nulled

- WHEN state has a dynamic setting value and config removes it
- WHEN update runs
- THEN the resource SHALL send that setting as `null` in the Put Settings request

#### Scenario: Template-injected mappings do not cause mapping update

- **GIVEN** an index is created with user-owned `mappings`
- **AND** a matching index template injects additional mapping `properties`, `dynamic_templates`, or other top-level mapping keys
- **WHEN** Terraform refreshes and plans the same index configuration
- **THEN** the resource SHALL treat the template-injected mapping content as non-drift and SHALL NOT call the Put Mapping API solely for those template-owned differences

### Requirement: Read (REQ-019–REQ-021)

On read, the resource SHALL parse `id` to extract the concrete index name, call the Get Index API with `flat_settings=true`, and if the index is not found (HTTP 404 or missing from response), SHALL remove the resource from state without error. When the index is found, the resource SHALL populate `concrete_name`, all aliases, `mappings`, `settings_raw`, and all individual setting attributes from the API response. For `mappings`, read SHALL preserve the user's prior mapping intent when the API response is a semantically equal superset caused by mappings injected by a matching index template. When state already contains a configured `name`, read SHALL preserve that configured value and SHALL NOT overwrite it with the concrete index name. When state does not contain `name`, read SHALL backfill `name` from the concrete index name.

#### Scenario: Index not found

- **WHEN** the Get Index API returns 404
- **THEN** the resource SHALL remove itself from state without error

#### Scenario: Date math name remains stable during read

- **WHEN** state already contains a configured date math expression in `name` and read refreshes the managed concrete index
- **THEN** `name` SHALL remain unchanged and `concrete_name` SHALL reflect the concrete index being managed

#### Scenario: Template-only mappings stay non-drifting

- **GIVEN** an index resource has no configured `mappings`
- **AND** a matching index template injects mappings into the created index
- **WHEN** read refreshes the index and Terraform plans the unchanged configuration
- **THEN** Terraform SHALL produce an empty plan for the index resource

#### Scenario: User-owned mappings tolerate template-injected extras

- **GIVEN** an index resource has configured `mappings`
- **AND** a matching index template injects additional mapping `properties`, `dynamic_templates`, or other top-level mapping keys
- **WHEN** read refreshes the index after create or during a later plan
- **THEN** Terraform SHALL NOT report a provider inconsistent-result error
- **AND** Terraform SHALL produce an empty plan for the unchanged configuration

### Requirement: Mappings plan modifier and semantic equality (REQ-022–REQ-024)

The `mappings` attribute SHALL use shared mapping comparison semantics for both semantic equality and replacement decisions. The comparison SHALL preserve existing mapped fields not present in config when those fields are user-owned and Elasticsearch would retain them after a field removal request. When a user-owned field is removed from config `mappings.properties`, the provider SHALL add a warning diagnostic and retain the field in the planned value or otherwise treat the retained field as semantically equal state. When a user-owned field's `type` changes between state and config, the provider SHALL require replacement. When `mappings.properties` is removed entirely from config while user-owned properties are present in state, the provider SHALL require replacement.

For mapping content injected by a matching index template, including additional `properties`, `dynamic_templates`, `_meta`, `runtime`, or other top-level mapping keys absent from user configuration, the resource SHALL treat the API value as a non-drifting superset of the user-owned mapping intent. The resource SHALL NOT require `lifecycle.ignore_changes = [mappings]` to avoid drift caused only by those template-injected mappings.

For `semantic_text` fields, Elasticsearch automatically enriches the stored mapping with a `model_settings` object (containing inference model configuration such as `dimensions`, `element_type`, `service`, `similarity`, and `task_type`) after index creation. When the field type in state and config is `semantic_text` and `model_settings` is present in state but absent from the config, the provider SHALL treat the enriched mapping as semantically equal to the configured mapping so that the plan matches the value Elasticsearch will return. When `model_settings` is explicitly specified in config, the config value SHALL be used as-is and SHALL NOT be overwritten by the state value.

#### Scenario: Field removed from config

- GIVEN state `mappings` contains user-owned field `foo` and config `mappings` does not
- WHEN plan runs
- THEN the plan SHALL retain `foo` in the planned `mappings` or treat the retained state value as semantically equal
- AND the provider SHALL add a warning diagnostic

#### Scenario: Field type changed

- GIVEN state `mappings` has user-owned field `foo` with `type: keyword` and config has `type: text`
- WHEN plan runs
- THEN the provider SHALL mark the resource for replacement

#### Scenario: semantic_text field without explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with server-enriched `model_settings`
- AND the config for that field does not specify `model_settings`
- WHEN plan runs
- THEN the provider SHALL treat the server-enriched `model_settings` as semantically equal to the configured field

#### Scenario: semantic_text field with explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with `model_settings`
- AND the config for that field also specifies `model_settings`
- WHEN plan runs
- THEN the provider SHALL use the config `model_settings` value and SHALL NOT overwrite it with the state value

#### Scenario: Template-injected dynamic templates are non-drift

- **GIVEN** a matching index template injects `dynamic_templates`
- **AND** the index resource configuration does not own those `dynamic_templates`
- **WHEN** Terraform compares refreshed mappings with prior user intent
- **THEN** the template-injected `dynamic_templates` SHALL be treated as non-drift

### Requirement: Settings mapping (REQ-025–REQ-027)

On create, the resource SHALL map each Terraform attribute to its corresponding Elasticsearch settings key (e.g. `mapping_total_fields_limit` → `mapping.total_fields.limit`) using dot-notation key conversion. Analysis settings (`analysis_analyzer`, `analysis_tokenizer`, `analysis_char_filter`, `analysis_filter`, `analysis_normalizer`) SHALL be parsed from JSON and nested under the `analysis` key in the create settings payload; these settings are applied only at index creation time and are not sent on update. When the deprecated `settings` block is also present, its `name`/`value` pairs SHALL be merged into the settings map; if any key conflicts with a dedicated attribute, the resource SHALL return an error diagnostic and SHALL not call the API.

#### Scenario: Duplicate setting detected

- GIVEN a setting is defined both via a dedicated attribute and in the deprecated `settings` block
- WHEN the provider builds the settings map
- THEN it SHALL return a "duplicate setting definition" error diagnostic and SHALL not call the Create or Put Settings API

### Requirement: Alias filter JSON mapping (REQ-028)

The `alias.filter` attribute SHALL be validated as a JSON string by the schema. When converting an alias to the API model, `filter` SHALL be unmarshalled from JSON into the API object when set. When reading aliases from the API, `filter` SHALL be serialized back to a JSON string and stored in state.

#### Scenario: Invalid alias filter JSON

- GIVEN an `alias.filter` value that is not valid JSON
- WHEN the schema validates the configuration
- THEN Terraform SHALL return a validation error before any API call is made

### Requirement: settings_raw computed attribute (REQ-029)

On every read, the resource SHALL serialize all index settings returned by the API to a normalized JSON string and store it in the computed `settings_raw` attribute.

#### Scenario: settings_raw populated on read

- GIVEN a successful Get Index API response
- WHEN the provider maps the response to state
- THEN `settings_raw` SHALL contain the JSON-serialized settings object from the API response

### Requirement: Opt-in adoption of existing indices via `use_existing`

The set of static settings compared during `use_existing` adoption SHALL be extended to include `sort.missing` and `sort.mode`. When these settings are explicitly set in the plan, the adoption flow SHALL compare them against the existing index's static settings and SHALL return an error diagnostic when they differ, consistent with the behavior for `sort.field` and `sort.order`.

#### Scenario: Adoption compares `sort.missing` against existing index

- **GIVEN** `use_existing = true` and an existing index where `index.sort.missing` is `["_last"]`
- **AND** the plan specifies `sort = [{ field = "date", missing = "_first" }]`
- **WHEN** create runs
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.missing` value
- **AND** SHALL NOT call any mutating API on the index

### Requirement: Nested `sort` attribute for per-field sort configuration (REQ-SORT-01)

The `elasticstack_elasticsearch_index` resource SHALL expose a new optional `sort` attribute as a `ListNestedAttribute`. Each element of the list SHALL represent one sort entry with the following nested attributes:

- `field` (required, string): The index field to sort by. Must have `doc_values` enabled (e.g. boolean, numeric, date, keyword).
- `order` (optional, string, allowed: `"asc"`, `"desc"`): The sort direction. Defaults to `"asc"` at the Elasticsearch level when not specified.
- `missing` (optional, string, allowed: `"_last"`, `"_first"`): How to treat documents that are missing the sort field. Defaults to `"_last"` at the Elasticsearch level when not specified.
- `mode` (optional, string, allowed: `"min"`, `"max"`): Which value to use when a sort field has multiple values. Defaults to `"min"` when order is `asc` and `"max"` when order is `desc` at the Elasticsearch level when not specified.

The `sort` attribute maps to the Elasticsearch static settings `index.sort.field`, `index.sort.order`, `index.sort.missing`, and `index.sort.mode`. Because these are immutable static settings, any change to the configured `sort` list SHALL require resource replacement, subject to the migration suppression rules in REQ-SORT-03.

The `sort` attribute and the deprecated `sort_field`/`sort_order` attributes SHALL be mutually exclusive. The schema SHALL enforce this with a `ConflictsWith` validator that produces a plan-time error when both `sort` and either `sort_field` or `sort_order` are set in the same configuration.

#### Scenario: Index created with nested sort attribute

- **GIVEN** a configuration with `sort = [{ field = "date", order = "desc", missing = "_last" }]`
- **WHEN** the resource is created
- **THEN** the Elasticsearch index SHALL be created with `index.sort.field = ["date"]`, `index.sort.order = ["desc"]`, and `index.sort.missing = ["_last"]`

#### Scenario: Multi-field sort preserves order

- **GIVEN** a configuration with `sort = [{ field = "date", order = "desc" }, { field = "username", order = "asc" }]`
- **WHEN** the resource is created
- **THEN** the Elasticsearch index SHALL be created with `index.sort.field = ["date", "username"]` and `index.sort.order = ["desc", "asc"]` in that order

#### Scenario: Mixing `sort` and `sort_field` is rejected at plan time

- **GIVEN** a configuration that sets both `sort` and `sort_field`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with a diagnostic before any API call is made

#### Scenario: Changing `sort` requires replace

- **GIVEN** an existing index managed with the `sort` attribute
- **WHEN** a configuration change modifies any entry in `sort` (e.g. changes `order`)
- **THEN** Terraform SHALL plan to destroy and recreate the resource

---

### Requirement: Deprecate `sort_field` and `sort_order` attributes (REQ-SORT-02)

The existing `sort_field` and `sort_order` attributes SHALL remain in the schema as deprecated optional attributes, still functioning as before. Both SHALL carry a `DeprecationMessage` directing users to use the `sort` attribute instead. Both SHALL carry `ConflictsWith` validators that produce a plan-time error when used alongside the new `sort` attribute.

#### Scenario: Deprecated attributes still work after provider upgrade

- **GIVEN** an existing configuration that uses `sort_field` and `sort_order`
- **WHEN** the provider is upgraded to the version containing this change
- **THEN** Terraform SHALL plan no changes and the resource SHALL remain under management without requiring migration

#### Scenario: Deprecation warning is surfaced

- **GIVEN** a configuration that uses `sort_field` or `sort_order`
- **WHEN** Terraform plans or applies
- **THEN** Terraform SHALL surface a deprecation warning for the attribute

---

### Requirement: Private-state-backed migration path from legacy to new `sort` attribute (REQ-SORT-03)

The resource SHALL store the ordered sort configuration from Elasticsearch in private state during every `Read` operation. This ordered sort configuration SHALL be stored under a private state key `"sort_config"` as a JSON-marshaled object containing ordered arrays for `fields`, `orders`, and optional per-position `missing`/`mode` values (as reported by Elasticsearch static settings).

When Terraform plans a configuration that:
1. Has `sort` as null in state (the resource was created using the deprecated attributes),
2. Has a non-null `sort` in the plan (the user is migrating to the new attribute),
3. And private state contains the ordered sort config from Elasticsearch,

the plan modifier on `sort` SHALL compare the plan's `sort[*].field` and `sort[*].order` (treating null `order` as `"asc"`) against the private state. The modifier SHALL also compare `sort[*].missing` and `sort[*].mode` using semantic normalization against existing index settings:

- Treat explicit defaults as equivalent to absent settings in both plan and Elasticsearch (`missing`: `"_last"`; `mode`: `"min"` when order is `asc`, `"max"` when order is `desc`).
- Compare values per position after order normalization.

When fields and orders match exactly (in order), and all planned `missing`/`mode` values are semantically equivalent to the existing index settings, the modifier SHALL suppress replace so users can migrate representations without destroying the index.

If any planned `sort[*].missing` or `sort[*].mode` value is not semantically equivalent to the existing index setting at the same position, replace SHALL be required.

If private state is absent (first `terraform apply` after provider upgrade before a `Read` has populated it), the modifier SHALL default to requiring replace. Users can avoid this by running `terraform refresh` before `terraform apply` after upgrading.

The deprecated `sort_field` and `sort_order` plan modifiers SHALL suppress replace for those attributes when `sort` is non-null in the plan (the new attribute's plan modifier owns the replace decision in that case).

#### Scenario: Migrating from legacy to new sort attribute does not replace the index

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **AND** the resource has been read at least once (private state is populated)
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc" }]`
- **THEN** Terraform SHALL NOT plan a destroy+recreate
- **AND** Terraform SHALL plan an in-place update (or no-change if no other attributes differ)

#### Scenario: Explicit default `missing`/`mode` values during migration do not require replace

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **AND** the existing index has no explicit `index.sort.missing` or `index.sort.mode` settings (Elasticsearch defaults apply)
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc", missing = "_last", mode = "max" }]`
- **THEN** Terraform SHALL NOT plan a destroy+recreate

#### Scenario: Non-equivalent `missing` or `mode` during migration requires replace

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc", missing = "_first" }]`
- **THEN** Terraform SHALL plan to destroy and recreate the resource

#### Scenario: First apply after upgrade forces replace when private state is absent

- **GIVEN** an existing index managed with `sort_field`/`sort_order` and no prior read since the provider was upgraded
- **WHEN** the configuration is changed to use `sort`
- **THEN** Terraform SHALL plan to destroy and recreate the resource (private state is not yet populated)

---

### Requirement: `sort.missing` and `sort.mode` in static settings and adoption (REQ-SORT-04)

The `sort.missing` and `sort.mode` Elasticsearch settings keys SHALL be added to `staticSettingsKeys`. The `use_existing` index adoption flow SHALL compare these settings against the existing index's static settings when they are explicitly set in the plan.

When the plan originates from the new `sort` `ListNestedAttribute`, the `compareStaticPlanAndES` function in `use_existing.go` SHALL compare `sort.field`, `sort.order`, `sort.missing`, and `sort.mode` as ordered string slices (using `stringSliceOrderedFromAny`). This preserves the order-significant semantics of the nested `sort` list defined in REQ-SORT-01.

When the plan originates from the deprecated `sort_field`/`sort_order` attributes, the `compareStaticPlanAndES` function in `use_existing.go` SHALL preserve the existing legacy behavior for `sort.field`: compare `sort.field` as an unordered set, while continuing to compare ordered per-position settings only where the plan shape preserves positional meaning.

#### Scenario: Adoption fails when nested `sort.field` order differs

- **GIVEN** `use_existing = true` and an existing index with `index.sort.field = ["date", "id"]`
- **AND** the plan specifies `sort = [{ field = "id" }, { field = "date" }]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.field` setting
- **AND** SHALL NOT call any mutating API

#### Scenario: Adoption preserves legacy unordered comparison for `sort_field`

- **GIVEN** `use_existing = true` and an existing index with `index.sort.field = ["date", "id"]`
- **AND** the plan specifies `sort_field = ["id", "date"]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL allow adoption without a `sort.field` mismatch based only on element order

#### Scenario: Adoption fails when `sort.missing` differs

- **GIVEN** `use_existing = true` and an existing index with `index.sort.missing = ["_last"]`
- **AND** the plan specifies `sort = [{ field = "date", missing = "_first" }]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.missing` setting
- **AND** SHALL NOT call any mutating API

---

### Requirement: Schema — sort attribute and deprecated sort_field/sort_order (REQ-SORT-05)

The `sort_field` and `sort_order` attributes SHALL remain in the schema as optional attributes but SHALL carry a `DeprecationMessage` directing users to the new `sort` attribute. The schema SHALL additionally expose the `sort` attribute as a `ListNestedAttribute` with nested `field` (required string), `order` (optional string), `missing` (optional string), and `mode` (optional string) attributes. The `sort` attribute and `sort_field`/`sort_order` SHALL be mutually exclusive; the schema SHALL enforce this with `ConflictsWith` validators.

```hcl
resource "elasticstack_elasticsearch_index" "example" {
  # Static settings (force new on change)
  sort_field = <optional, set(string), DEPRECATED — use sort>  # force new
  sort_order = <optional, list(string), DEPRECATED — use sort> # force new

  # Replaces sort_field and sort_order
  sort = <optional, list(object)> {   # force new (with migration suppression — see REQ-SORT-03)
    field   = <required, string>
    order   = <optional, string>  # allowed: "asc", "desc"
    missing = <optional, string>  # allowed: "_last", "_first"
    mode    = <optional, string>  # allowed: "min", "max"
  }
}
```

#### Scenario: Deprecated attributes emit deprecation warnings

- **GIVEN** a Terraform configuration that uses `sort_field` or `sort_order`
- **WHEN** Terraform validates or applies the configuration
- **THEN** a deprecation warning SHALL be surfaced for each deprecated attribute

---

