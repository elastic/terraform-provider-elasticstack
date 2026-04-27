## 1. Model layer

- [ ] 1.1 Add `DataStreamOptions` and `FailureStoreOptions` structs to `internal/models/models.go`
- [ ] 1.2 Add `DataStreamOptions *DataStreamOptions` field (JSON tag `data_stream_options,omitempty`) to `models.Template`

## 2. Schema changes — resource

- [ ] 2.1 Add `MinSupportedDataStreamOptionsVersion` constant (`9.0.0`) alongside the existing version constants in `internal/elasticsearch/index/template.go`
- [ ] 2.2 Add the `data_stream_options` block to the `template` block schema in `ResourceTemplate()`:
  - `data_stream_options`: `TypeList`, optional, `MaxItems: 1`
  - `failure_store`: `TypeList`, optional, `MaxItems: 1`, inside `data_stream_options`
  - `enabled`: `TypeBool`, required, inside `failure_store`
- [ ] 2.3 Extend `expandTemplate` to read `data_stream_options` from the Terraform config and populate `templ.DataStreamOptions`
- [ ] 2.4 Add a version gate in `resourceIndexTemplatePut` that returns an error diagnostic when `data_stream_options` is set and the server version is below `9.0.0`
- [ ] 2.5 Extend `flattenTemplateData` to write `data_stream_options` back to state when `template.DataStreamOptions != nil`

## 3. Schema changes — data source

- [ ] 3.1 Add the same `data_stream_options` block definition (read-only) to the `template` block schema in `DataSourceTemplate()` in `internal/elasticsearch/index/template_data_source.go`
- [ ] 3.2 Extend the data source flatten path to populate `data_stream_options` state from the API response

## 4. Spec update

- [ ] 4.1 Create `openspec/changes/index-template-data-stream-options/specs/elasticsearch-index-template/spec.md` with the delta spec for the new requirements (REQ-032–REQ-037)

## 5. Tests

- [ ] 5.1 Add unit tests for `expandTemplate` covering `data_stream_options` populated and absent
- [ ] 5.2 Add unit tests for `flattenTemplateData` covering `data_stream_options` populated and absent
- [ ] 5.3 Add acceptance test `TestAccResourceIndexTemplateDataStreamOptions` with testdata configs:
  - `create/index_template.tf` — data stream template with `data_stream_options.failure_store.enabled = true`
  - `update/index_template.tf` — same template with `enabled = false`
  - `remove/index_template.tf` — template without `data_stream_options` block
- [ ] 5.4 Add acceptance test for the data source reading `data_stream_options` from an existing template

## 6. Documentation

- [ ] 6.1 Regenerate `docs/resources/elasticsearch_index_template.md` to include the new `data_stream_options` block
- [ ] 6.2 Regenerate `docs/data-sources/elasticsearch_index_template.md` similarly
