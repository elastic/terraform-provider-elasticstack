## 1. Migrate Remaining 35 Processors

- [x] 1.1 Migrate `bytes` processor to PF
- [x] 1.2 Migrate `circle` processor to PF
- [x] 1.3 Migrate `community_id` processor to PF
- [x] 1.4 Migrate `convert` processor to PF
- [x] 1.5 Migrate `csv` processor to PF
- [x] 1.6 Migrate `date` processor to PF
- [x] 1.7 Migrate `date_index_name` processor to PF
- [x] 1.8 Migrate `dissect` processor to PF
- [x] 1.9 Migrate `dot_expander` processor to PF
- [x] 1.10 Migrate `enrich` processor to PF
- [x] 1.11 Migrate `fail` processor to PF
- [x] 1.12 Migrate `fingerprint` processor to PF
- [x] 1.13 Migrate `grok` processor to PF
- [x] 1.14 Migrate `gsub` processor to PF
- [x] 1.15 Migrate `html_strip` processor to PF
- [x] 1.16 Migrate `inference` processor to PF
- [x] 1.17 Migrate `join` processor to PF
- [x] 1.18 Migrate `json` processor to PF
- [x] 1.19 Migrate `kv` processor to PF
- [x] 1.20 Migrate `lowercase` processor to PF
- [x] 1.21 Migrate `network_direction` processor to PF
- [x] 1.22 Migrate `pipeline` processor to PF
- [x] 1.23 Migrate `registered_domain` processor to PF
- [x] 1.24 Migrate `remove` processor to PF
- [x] 1.25 Migrate `rename` processor to PF
- [x] 1.26 Migrate `reroute` processor to PF
- [x] 1.27 Migrate `set` processor to PF
- [x] 1.28 Migrate `set_security_user` processor to PF
- [x] 1.29 Migrate `sort` processor to PF
- [x] 1.30 Migrate `split` processor to PF
- [x] 1.31 Migrate `trim` processor to PF
- [x] 1.32 Migrate `uppercase` processor to PF
- [x] 1.33 Migrate `uri_parts` processor to PF
- [x] 1.34 Migrate `urldecode` processor to PF
- [x] 1.35 Migrate `geoip` processor to PF (adds common fields: description, if, ignore_failure, on_failure, tag)
- [x] 1.36 Migrate `user_agent` processor to PF (adds common fields: description, if, ignore_failure, on_failure, tag)

## 2. Provider Wiring

- [x] 2.1 Register all 35 new constructors in `provider/plugin_framework.go`
- [x] 2.2 Remove all 35 old SDK registrations from `provider/provider.go`

## 3. Cleanup

- [x] 3.1 Delete old SDK data source implementation files (`processor_*_data_source.go`) for all 39 processors
- [x] 3.2 Delete old SDK data source test files (`processor_*_data_source_test.go`) for all 39 processors
- [x] 3.3 Delete `internal/elasticsearch/ingest/commons_test.go` if no longer needed — **N/A**, still used by `pipeline_test.go`
- [x] 3.4 Move remaining processor structs from `internal/models/ingest.go` to `internal/elasticsearch/ingest/processor_models.go`
- [x] 3.5 Delete processor structs from `internal/models/ingest.go`

## 4. Verification

- [x] 4.1 Run `make build` and verify no compilation errors
- [x] 4.2 Run full acceptance test suite for all migrated processor data sources
- [x] 4.3 Run `make check-openspec` and verify the change passes validation
