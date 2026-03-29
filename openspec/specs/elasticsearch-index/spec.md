# `elasticstack_elasticsearch_index` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/index`

## Purpose

Define schema and behavior for the Elasticsearch index resource: API usage, identity/import, connection, lifecycle (static vs dynamic settings, deletion protection), mappings plan modifier, alias management, and settings mapping between Terraform state and the Elasticsearch indices API.

## Schema

```hcl
resource "elasticstack_elasticsearch_index" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<index_name>
  name = <required, string> # force new; 1–255 chars, lowercase alphanumeric + selected punctuation, cannot start with -, _, +

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

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<index_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `name` and store it in state.

#### Scenario: Id set on create

- GIVEN a successful Create Index API response
- WHEN the provider stores the result in state
- THEN `id` SHALL be set to `<cluster_uuid>/<index_name>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import by accepting an `id` value directly via `ImportStatePassthroughID`, persisting the imported `id` to state without validation at import time. Read and delete operations SHALL parse `id` in the format `<cluster_uuid>/<index_name>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN an import command with a composite `id`
- WHEN import completes
- THEN the `id` SHALL be stored in state for subsequent read and plan operations

### Requirement: Lifecycle — static settings require replacement (REQ-009)

Changing the `name` attribute SHALL require resource replacement. Changing any static index setting (`number_of_shards`, `number_of_routing_shards`, `codec`, `routing_partition_size`, `load_fixed_bitset_filters_eagerly`, `shard_check_on_startup`, `sort_field`, `sort_order`, `mapping_coerce`) SHALL also require resource replacement, because Elasticsearch does not allow these to be changed on an existing index.

#### Scenario: Name change forces new

- GIVEN an existing index resource
- WHEN the `name` attribute is changed in configuration
- THEN Terraform SHALL plan to destroy and recreate the resource

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

On create, the resource SHALL build an API model from the plan (including settings, mappings, and aliases), set `id` from the cluster UUID and `name`, and submit a Create Index request with the configured `wait_for_active_shards`, `master_timeout`, and `timeout` parameters. After a successful create, the resource SHALL perform a read to refresh all computed attributes in state.

#### Scenario: Serverless — master_timeout and wait_for_active_shards omitted

- GIVEN the Elasticsearch server flavor is `serverless`
- WHEN a create request is issued
- THEN `master_timeout` and `wait_for_active_shards` SHALL be omitted from the API call parameters

### Requirement: Update flow (REQ-015–REQ-018)

On update, the resource SHALL only call the relevant update APIs when the corresponding values have changed. Alias changes SHALL be applied by deleting aliases removed from config (via Delete Alias API) and upserting all aliases present in plan (via Put Alias API). Dynamic setting changes SHALL be applied by calling the Put Settings API with the diff, setting removed dynamic settings to `null` in the request. Mapping changes SHALL be applied by calling the Put Mapping API when `mappings` has semantically changed. After all updates, the resource SHALL perform a read to refresh state.

#### Scenario: Removed alias is deleted

- GIVEN an alias exists in state but is absent from the plan
- WHEN update runs
- THEN the resource SHALL call the Delete Alias API for that alias

#### Scenario: Removed dynamic setting set to null

- GIVEN a dynamic setting is present in state but absent from the plan
- WHEN update runs
- THEN the resource SHALL send that setting as `null` in the Put Settings request

### Requirement: Read (REQ-019–REQ-021)

On read, the resource SHALL parse `id` to extract the index name, call the Get Index API with `flat_settings=true`, and if the index is not found (HTTP 404 or missing from response), SHALL remove the resource from state without error. When the index is found, the resource SHALL populate `name`, all aliases, `mappings`, `settings_raw`, and all individual setting attributes from the API response.

#### Scenario: Index not found

- GIVEN the Get Index API returns 404 or the index name is absent from the response
- WHEN read runs
- THEN the resource SHALL be removed from state and no error diagnostic SHALL be added

### Requirement: Mappings plan modifier (REQ-022–REQ-024)

The `mappings` attribute SHALL use a custom plan modifier that preserves existing mapped fields not present in config, because Elasticsearch ignores field removal requests. When a field is removed from config `mappings.properties`, the plan modifier SHALL add a warning diagnostic and retain the field in the planned value. When a field's `type` changes between state and config, the plan modifier SHALL require replacement. When `mappings.properties` is removed entirely from config while present in state, the plan modifier SHALL require replacement.

#### Scenario: Field removed from config

- GIVEN state `mappings` contains field `foo` and config `mappings` does not
- WHEN plan runs
- THEN the plan SHALL retain `foo` in the planned `mappings` and SHALL add a warning diagnostic

#### Scenario: Field type changed

- GIVEN state `mappings` has field `foo` with `type: keyword` and config has `type: text`
- WHEN plan runs
- THEN the plan modifier SHALL mark the resource for replacement

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
