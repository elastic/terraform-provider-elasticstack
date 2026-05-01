## 1. Shared Base Infrastructure

- [ ] 1.1 Create `internal/elasticsearch/ingest/processor_datasource_base.go` with `ProcessorModel` interface, generic `processorDataSource[T]` struct, and `marshalAndHash()` helper
- [ ] 1.2 Create `internal/elasticsearch/ingest/processor_common.go` with `CommonProcessorModel`, `CommonProcessorSchemaAttributes()`, and `appendCommonFields()` helper
- [ ] 1.3 Create `internal/elasticsearch/ingest/processor_models.go` with local inner structs for `drop`, `append`, `script`, `foreach` (keep `json` tags identical to existing `models.ProcessorX`)

## 2. Representative Processor Migrations

- [ ] 2.1 Create `internal/elasticsearch/ingest/processor_drop_data_source.go` (PF schema + model + `MarshalBody()` + constructor)
- [ ] 2.2 Create `internal/elasticsearch/ingest/processor_append_data_source.go` (PF schema + model + `MarshalBody()` + constructor)
- [ ] 2.3 Create `internal/elasticsearch/ingest/processor_script_data_source.go` (PF schema + model + `MarshalBody()` + constructor; include `ExactlyOneOf` validator for `script_id` vs `source`)
- [ ] 2.4 Create `internal/elasticsearch/ingest/processor_foreach_data_source.go` (PF schema + model + `MarshalBody()` + constructor; handle `processor` JSON string → map parsing)

## 3. Provider Wiring

- [ ] 3.1 Register the 4 new constructors in `provider/plugin_framework.go`
- [ ] 3.2 Remove the 4 old SDK registrations from `provider/provider.go` `DataSourcesMap`

## 4. Verification

- [ ] 4.1 Run `make build` and verify no compilation errors
- [ ] 4.2 Run targeted acceptance tests for the 4 migrated processors
- [ ] 4.3 Run `make check-openspec` and verify the change passes validation
