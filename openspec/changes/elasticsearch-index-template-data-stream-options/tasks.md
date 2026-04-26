## 1. Model extension

- [ ] 1.1 Add `FailureStoreOptions` struct to `internal/models/models.go` with `Enabled *bool \`json:"enabled,omitempty"\``
- [ ] 1.2 Add `DataStreamOptions` struct to `internal/models/models.go` with `FailureStore *FailureStoreOptions \`json:"failure_store,omitempty"\``
- [ ] 1.3 Add `DataStreamOptions *DataStreamOptions \`json:"data_stream_options,omitempty"\`` field to the `Template` struct in `internal/models/models.go`

## 2. Resource schema and CRUD wiring

- [ ] 2.1 Add the `data_stream_options` optional `TypeList/MaxItems:1` block to the `template` elem schema in `ResourceTemplate()` inside `internal/elasticsearch/index/template.go`, with a nested `failure_store` block containing an `enabled` optional bool attribute
- [ ] 2.2 Extend `expandTemplate` in `internal/elasticsearch/index/template.go` to read `data_stream_options` from the decoded template map and populate `templ.DataStreamOptions`
- [ ] 2.3 Extend `flattenTemplateData` in `internal/elasticsearch/index/template.go` to write `data_stream_options` into the Terraform state map when `template.DataStreamOptions` is non-nil

## 3. Data source schema wiring

- [ ] 3.1 Add the same `data_stream_options` optional `TypeList/MaxItems:1` block to the `template` elem schema in the data source definition inside `internal/elasticsearch/index/template_data_source.go`

## 4. Verification and documentation

- [ ] 4.1 Add acceptance test config files under `internal/elasticsearch/index/testdata/` for a create step, update step (removing `data_stream_options`), and an optional data-stream-options-set step for `elasticstack_elasticsearch_index_template`
- [ ] 4.2 Add or update the acceptance test case(s) in `internal/elasticsearch/index/template_test.go` to cover `data_stream_options` create, read, update (add and remove), and state round-trip
- [ ] 4.3 Regenerate `docs/resources/elasticsearch_index_template.md` and `docs/data-sources/elasticsearch_index_template.md` by running `make docs`
- [ ] 4.4 Run `make build` and resolve any compilation errors
