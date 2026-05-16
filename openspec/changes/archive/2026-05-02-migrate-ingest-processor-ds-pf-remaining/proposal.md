## Why

The shared generic base introduced in PR #1 (`migrate-ingest-processor-ds-pf-shared-base`) proves the pattern. Now the remaining 35 processor data sources must be migrated from the Terraform Plugin SDK to the Plugin Framework to complete the migration. Additionally, two processors (`geoip`, `user_agent`) currently omit common processor fields that Elasticsearch supports — this is corrected as part of the migration.

## What Changes

- **Migrate** the remaining 35 processor data sources to Plugin Framework using the established shared base pattern:
  - `bytes`, `circle`, `community_id`, `convert`, `csv`, `date`, `date_index_name`, `dissect`, `dot_expander`, `enrich`, `fail`, `fingerprint`, `grok`, `gsub`, `html_strip`, `inference`, `join`, `json`, `kv`, `lowercase`, `network_direction`, `pipeline`, `registered_domain`, `remove`, `rename`, `reroute`, `set`, `set_security_user`, `sort`, `split`, `trim`, `uppercase`, `uri_parts`, `urldecode`
- **Migrate** `geoip` and `user_agent` with added common processor fields (`description`, `if`, `ignore_failure`, `on_failure`, `tag`)
- **Register** all 35 new constructors in `provider/plugin_framework.go`
- **Remove** all 35 old SDK registrations from `provider/provider.go`
- **Delete** old SDK data source implementation files (`processor_*_data_source.go`)
- **Delete** old SDK data source test files (`processor_*_data_source_test.go`)
- **Delete** `commons_test.go` if no longer referenced
- **Move** processor model structs from `internal/models/ingest.go` to `internal/elasticsearch/ingest/processor_models.go`

## Capabilities

### New Capabilities

_None — this is an implementation migration._

### Modified Capabilities

The following existing capabilities are migrated from Plugin SDK to Plugin Framework with identical behavior (no requirement changes):

- `elasticsearch-ingest-processor-bytes`
- `elasticsearch-ingest-processor-circle`
- `elasticsearch-ingest-processor-community-id`
- `elasticsearch-ingest-processor-convert`
- `elasticsearch-ingest-processor-csv`
- `elasticsearch-ingest-processor-date`
- `elasticsearch-ingest-processor-date-index-name`
- `elasticsearch-ingest-processor-dissect`
- `elasticsearch-ingest-processor-dot-expander`
- `elasticsearch-ingest-processor-enrich`
- `elasticsearch-ingest-processor-fail`
- `elasticsearch-ingest-processor-fingerprint`
- `elasticsearch-ingest-processor-foreach` (already migrated in PR #1, will be cleaned up here)
- `elasticsearch-ingest-processor-grok`
- `elasticsearch-ingest-processor-gsub`
- `elasticsearch-ingest-processor-html-strip`
- `elasticsearch-ingest-processor-inference`
- `elasticsearch-ingest-processor-join`
- `elasticsearch-ingest-processor-json`
- `elasticsearch-ingest-processor-kv`
- `elasticsearch-ingest-processor-lowercase`
- `elasticsearch-ingest-processor-network-direction`
- `elasticsearch-ingest-processor-pipeline`
- `elasticsearch-ingest-processor-registered-domain`
- `elasticsearch-ingest-processor-remove`
- `elasticsearch-ingest-processor-rename`
- `elasticsearch-ingest-processor-reroute`
- `elasticsearch-ingest-processor-script` (already migrated in PR #1, will be cleaned up here)
- `elasticsearch-ingest-processor-set`
- `elasticsearch-ingest-processor-set-security-user`
- `elasticsearch-ingest-processor-sort`
- `elasticsearch-ingest-processor-split`
- `elasticsearch-ingest-processor-trim`
- `elasticsearch-ingest-processor-uppercase`
- `elasticsearch-ingest-processor-uri-parts`
- `elasticsearch-ingest-processor-urldecode`

The following capabilities have **true requirement changes** (new optional attributes):

- `elasticsearch-ingest-processor-geoip`
- `elasticsearch-ingest-processor-user-agent`

## Impact

- `internal/elasticsearch/ingest/`: +35 new PF files, deletion of 39 old SDK files and tests, +local model structs moved
- `provider/plugin_framework.go`: +35 data source registrations
- `provider/provider.go`: -35 data source registrations
- `internal/models/ingest.go`: processor structs removed
- Acceptance tests: all existing tests transparently exercise PF implementation (already use `ProtoV6ProviderFactories`)
