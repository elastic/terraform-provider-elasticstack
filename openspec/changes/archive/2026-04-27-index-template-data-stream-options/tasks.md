## 1. Model layer

- [x] 1.1 Add `DataStreamOptions`, `FailureStoreOptions`, and `FailureStoreLifecycle` struct types to `internal/models/models.go`
- [x] 1.2 Add `DataStreamOptions *DataStreamOptions` field tagged `json:"data_stream_options,omitempty"` to the `models.Template` struct

## 2. Schema and expand/flatten

- [x] 2.1 Add `data_stream_options` block to the `template` schema in `internal/elasticsearch/index/template.go`, with a nested `failure_store` block containing `enabled` (required bool) and an optional `lifecycle` block containing `data_retention` (required string)
- [x] 2.2 Add a minimum server version constant `MinSupportedDataStreamOptionsVersion = "9.1.0"` (following the pattern of `MinSupportedIgnoreMissingComponentTemplateVersion`)
- [x] 2.3 Update `resourceIndexTemplatePut` to version-gate `data_stream_options`: if the attribute is configured and the server version is below `9.1.0`, return an error diagnostic without calling the Put API
- [x] 2.4 Update `expandTemplate` to read `data_stream_options` from the Terraform config map and populate `templ.DataStreamOptions` accordingly
- [x] 2.5 Update `flattenTemplateData` to convert a non-nil `template.DataStreamOptions` into the nested HCL block representation in state

## 3. Acceptance tests

- [x] 3.1 Add test data fixtures under `internal/elasticsearch/index/testdata/` for a new acceptance test covering `data_stream_options` (at minimum: create with `failure_store.enabled = true`, read-back, and update)
- [x] 3.2 Add the acceptance test function(s) to `internal/elasticsearch/index/template_test.go` that exercise `data_stream_options` behavior end-to-end

## 4. Documentation

- [x] 4.1 Regenerate `docs/resources/elasticsearch_index_template.md` using `make generate-docs` (or equivalent) to reflect the new `data_stream_options` block and its attributes
