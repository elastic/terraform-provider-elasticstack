# `elasticstack_elasticsearch_indices` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/index/indices`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_indices` data source, which retrieves information about one or more existing Elasticsearch indices matching a target pattern. The data source calls the Elasticsearch Get index API with an optional target expression (index name, alias, wildcard, or `_all`) and populates a list of index objects covering settings, mappings, and aliases. When no target is specified, it defaults to `*` (all indices).

## Schema

```hcl
data "elasticstack_elasticsearch_indices" "example" {
  # Optional filter; omit or use "*" / "_all" to return all indices
  target = <optional, string>

  # Computed top-level attributes
  id = <computed, string>  # set to the effective target string used in the API call

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }

  # Computed: list of matching indices
  indices = <computed, list(object)> [
    {
      id   = <computed, string>  # internal identifier (cluster_uuid/name)
      name = <required, string>  # index name (1–255 chars, validated)

      # Static settings (set only on creation)
      number_of_shards               = <optional, int>
      number_of_routing_shards       = <optional, int>
      codec                          = <optional, string>  # "best_compression"
      routing_partition_size         = <optional, int>
      load_fixed_bitset_filters_eagerly = <optional, bool>
      shard_check_on_startup         = <optional, string>  # "false" | "true" | "checksum"
      sort_field                     = <optional, set(string)>
      sort_order                     = <optional, list(string)>  # ordered; allows duplicates
      mapping_coerce                 = <optional, bool>

      # Dynamic settings (can change at runtime)
      number_of_replicas             = <optional+computed, int>
      auto_expand_replicas           = <optional, string>
      search_idle_after              = <optional, string>
      refresh_interval               = <optional, string>
      max_result_window              = <optional, int>
      max_inner_result_window        = <optional, int>
      max_rescore_window             = <optional, int>
      max_docvalue_fields_search     = <optional, int>
      max_script_fields              = <optional, int>
      max_ngram_diff                 = <optional, int>
      max_shingle_diff               = <optional, int>
      max_refresh_listeners          = <optional, int>
      analyze_max_token_count        = <optional, int>
      highlight_max_analyzed_offset  = <optional, int>
      max_terms_count                = <optional, int>
      max_regex_length               = <optional, int>
      query_default_field            = <optional, set(string)>
      routing_allocation_enable      = <optional, string>  # "all" | "primaries" | "new_primaries" | "none"
      routing_rebalance_enable       = <optional, string>  # "all" | "primaries" | "replicas" | "none"
      gc_deletes                     = <optional, string>
      blocks_read_only               = <optional, bool>
      blocks_read_only_allow_delete  = <optional, bool>
      blocks_read                    = <optional, bool>
      blocks_write                   = <optional, bool>
      blocks_metadata                = <optional, bool>
      default_pipeline               = <optional, string>
      final_pipeline                 = <optional, string>
      unassigned_node_left_delayed_timeout = <optional, string>

      # Slowlog thresholds and levels
      search_slowlog_threshold_query_warn    = <optional, string>
      search_slowlog_threshold_query_info    = <optional, string>
      search_slowlog_threshold_query_debug   = <optional, string>
      search_slowlog_threshold_query_trace   = <optional, string>
      search_slowlog_threshold_fetch_warn    = <optional, string>
      search_slowlog_threshold_fetch_info    = <optional, string>
      search_slowlog_threshold_fetch_debug   = <optional, string>
      search_slowlog_threshold_fetch_trace   = <optional, string>
      search_slowlog_level                   = <optional, string>  # "warn" | "info" | "debug" | "trace"
      indexing_slowlog_threshold_index_warn  = <optional, string>
      indexing_slowlog_threshold_index_info  = <optional, string>
      indexing_slowlog_threshold_index_debug = <optional, string>
      indexing_slowlog_threshold_index_trace = <optional, string>
      indexing_slowlog_level                 = <optional, string>  # "warn" | "info" | "debug" | "trace"
      indexing_slowlog_source                = <optional, string>

      # Analysis settings (JSON normalized strings; not populated from API flat-settings)
      analysis_analyzer    = <optional, json string>
      analysis_tokenizer   = <optional, json string>
      analysis_char_filter = <optional, json string>
      analysis_filter      = <optional, json string>
      analysis_normalizer  = <optional, json string>

      # Operational settings
      deletion_protection      = <optional+computed, bool>
      wait_for_active_shards   = <optional+computed, string>
      master_timeout           = <optional+computed, string>  # duration type
      timeout                  = <optional+computed, string>  # duration type

      # Mappings and raw settings (JSON normalized strings)
      mappings     = <optional+computed, json string>
      settings_raw = <computed, json string>  # all flat settings from the cluster

      # Aliases
      alias = <optional, set(object)> [
        {
          name           = <required, string>
          filter         = <optional, json string>
          index_routing  = <optional+computed, string>
          is_hidden      = <optional+computed, bool>
          is_write_index = <optional+computed, bool>
          routing        = <optional+computed, string>
          search_routing = <optional+computed, string>
        }
      ]
    }
  ]
}
```

## Requirements

### Requirement: Read API (REQ-001)

The data source SHALL use the Elasticsearch Get index API (`esClient.Indices.Get`) to retrieve index information ([Get index API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-index.html)). The API SHALL be called with flat settings enabled (`WithFlatSettings(true)`) so that all settings are returned as dot-notation keys.

#### Scenario: API call with flat settings

- GIVEN the data source is read
- WHEN the provider calls Elasticsearch
- THEN it SHALL call `Indices.Get` with `flat_settings=true`

### Requirement: API error surfacing (REQ-002)

When the Elasticsearch Get index API returns a non-success status (other than 404), the data source SHALL surface the error in Terraform diagnostics and SHALL NOT update state.

#### Scenario: Non-success API response

- GIVEN the Elasticsearch API returns a non-404 error
- WHEN the data source reads indices
- THEN the error SHALL appear in Terraform diagnostics and state SHALL NOT be updated

### Requirement: Target defaulting (REQ-003)

When `target` is null or an empty string, the data source SHALL use `"*"` as the effective target when calling the Elasticsearch API. The computed `id` SHALL be set to the effective target string (i.e., `"*"` when `target` is omitted).

#### Scenario: Omitted target

- GIVEN `target` is not set in the configuration
- WHEN the data source reads indices
- THEN the API SHALL be called with target `"*"` and `id` SHALL equal `"*"`

#### Scenario: Explicit wildcard target

- GIVEN `target` is set to `"*"` or `"_all"` or a comma-separated pattern
- WHEN the data source reads indices
- THEN the API SHALL be called with that exact value and `id` SHALL equal that value

### Requirement: Empty result on 404 (REQ-004)

When the Elasticsearch API returns HTTP 404 (no indices match the target), the data source SHALL return an empty `indices` list and SHALL NOT surface a diagnostic error.

#### Scenario: No matching indices

- GIVEN no indices match the target pattern
- WHEN the API returns 404
- THEN the data source SHALL return an empty `indices` list without an error

### Requirement: Identity (REQ-005)

The data source SHALL expose a computed `id` set to the effective target string used in the API call. When `target` is provided, `id` SHALL equal `target`. When `target` is omitted, `id` SHALL equal `"*"`.

#### Scenario: id reflects target

- GIVEN `target` is set to `".security-*"`
- WHEN the data source reads
- THEN `id` SHALL equal `".security-*"`

### Requirement: Index list population (REQ-006)

For each index returned by the Elasticsearch API, the data source SHALL populate one entry in the computed `indices` list. Each entry SHALL have its `name` set to the index name as returned by the API.

#### Scenario: Multiple indices returned

- GIVEN a wildcard target matching two indices
- WHEN the data source reads
- THEN `indices` SHALL contain exactly one entry per matching index

### Requirement: Settings mapping from flat settings (REQ-007)

For each index, the data source SHALL map Elasticsearch flat settings (prefixed `index.<key>`) to the corresponding typed state attributes. String settings SHALL be stored as strings; boolean settings (which Elasticsearch may return as the string `"true"` or `"false"`) SHALL be parsed to booleans; integer settings (which Elasticsearch may return as strings) SHALL be parsed to int64 values. If type conversion fails, the data source SHALL return an error diagnostic.

#### Scenario: Boolean setting returned as string

- GIVEN Elasticsearch returns `"index.blocks.read"` as the string `"false"`
- WHEN the data source maps settings to state
- THEN `blocks_read` SHALL be stored as the boolean `false`

#### Scenario: Integer setting returned as string

- GIVEN Elasticsearch returns `"index.number_of_replicas"` as the string `"1"`
- WHEN the data source maps settings to state
- THEN `number_of_replicas` SHALL be stored as the integer `1`

### Requirement: Raw settings always populated (REQ-008)

For each index, the data source SHALL serialize the complete flat settings map from the Elasticsearch API response to JSON and store it in the computed `settings_raw` attribute as a normalized JSON string.

#### Scenario: settings_raw reflects API response

- GIVEN an index with known settings
- WHEN the data source reads
- THEN `settings_raw` SHALL contain a JSON object including all flat settings returned by the API

### Requirement: Mappings population (REQ-009)

When an index has mappings in the API response, the data source SHALL serialize the mappings to a normalized JSON string and store it in the computed `mappings` attribute. When an index has no mappings, `mappings` SHALL be null.

#### Scenario: Index with mappings

- GIVEN an index with an explicit mappings definition
- WHEN the data source reads
- THEN `mappings` SHALL be a normalized JSON string representation of those mappings

#### Scenario: Index with no mappings

- GIVEN an index with no mappings
- WHEN the data source reads
- THEN `mappings` SHALL be null

### Requirement: Analysis settings not populated from flat settings (REQ-010)

The data source SHALL NOT populate `analysis_analyzer`, `analysis_tokenizer`, `analysis_char_filter`, `analysis_filter`, or `analysis_normalizer` from the Elasticsearch flat-settings API response. Analysis configuration lives under the `index.analysis.*` key namespace, which is outside the static and dynamic settings key sets iterated by the provider. These attributes MAY be present in `settings_raw`.

#### Scenario: Analysis settings absent from typed attributes

- GIVEN an index with a custom analyzer configured
- WHEN the data source reads
- THEN `analysis_analyzer` and related analysis attributes SHALL remain null (not populated from the API)
- AND the analysis configuration SHALL be visible in `settings_raw`

### Requirement: Alias population (REQ-011)

For each index, the data source SHALL populate the `alias` set with one entry per alias returned by the Elasticsearch API. Each alias entry SHALL include `name`, `index_routing`, `is_hidden`, `is_write_index`, `routing`, and `search_routing` from the API response. When an alias has a `filter`, the data source SHALL serialize it to a normalized JSON string; when absent, `filter` SHALL be null.

#### Scenario: Index with aliases

- GIVEN an index with one or more aliases
- WHEN the data source reads
- THEN `alias` SHALL contain one entry per alias with all fields correctly populated

#### Scenario: Alias filter serialization

- GIVEN an alias with a filter query
- WHEN the data source reads
- THEN the alias `filter` SHALL be a normalized JSON string of the filter

### Requirement: Elasticsearch connection (REQ-012–REQ-013)

The data source SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured on the data source, the data source SHALL use that connection to create an Elasticsearch client for all API calls of that instance.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set on the data source
- WHEN any API call runs for that instance
- THEN the client SHALL be built from that block

### Requirement: Index name validation (REQ-014)

Within each `indices` entry, the `name` attribute SHALL be validated to enforce Elasticsearch index naming rules: length between 1 and 255 characters, not equal to `"."` or `".."`, not starting with `-`, `_`, or `+`, and containing only lowercase alphanumeric characters and the following punctuation: `!`, `$`, `%`, `&`, `'`, `(`, `)`, `+`, `.`, `;`, `=`, `@`, `[`, `]`, `^`, `{`, `}`, `~`, `_`, `-`.

#### Scenario: Invalid index name rejected

- GIVEN an `indices` entry with a name starting with `_`
- WHEN the schema is validated
- THEN the provider SHALL return a validation error

### Requirement: List attribute ordering (REQ-015)

The `sort_order` attribute within each `indices` entry SHALL be modeled as a `list(string)` (preserving order and allowing duplicate values) rather than a set, to support cases where the same direction appears for multiple sort fields.

#### Scenario: Duplicate sort directions preserved

- GIVEN an index with `sort_order` containing duplicate values such as `["asc", "asc"]`
- WHEN the data source reads and maps settings
- THEN the list order and duplicates SHALL be preserved in state
