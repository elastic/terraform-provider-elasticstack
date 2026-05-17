## 1. Internal model

- [ ] 1.1 Add `AllowAutoCreate *bool` with `json:"allow_auto_create,omitempty"` to `IndexTemplate` in `internal/models/models.go`.

## 2. Terraform model

- [ ] 2.1 Add `AllowAutoCreate types.Bool` field (tfsdk tag `allow_auto_create`) to `Model` in `internal/elasticsearch/index/template/models.go`.

## 3. Resource schema

- [ ] 3.1 Add `allow_auto_create` as `schema.BoolAttribute{Optional: true, MarkdownDescription: descAllowAutoCreate}` to the top-level `Attributes` map in `resourceSchema()` in `internal/elasticsearch/index/template/schema.go`.

## 4. Data source schema

- [ ] 4.1 Add `allow_auto_create` as `dschema.BoolAttribute{Computed: true, MarkdownDescription: descAllowAutoCreate}` to the top-level `Attributes` map in `getDataSourceSchema()` in `internal/elasticsearch/index/template/data_source_schema.go`.

## 5. Descriptions

- [ ] 5.1 Add `descAllowAutoCreate` constant to `internal/elasticsearch/index/template/descriptions.go` with text: "If true, index auto-creation is allowed for matching indices. If false, auto-creation is disabled for matching indices. When unset, the cluster-level ``action.auto_create_index`` setting applies.".

## 6. Expand

- [ ] 6.1 In `(Model).toAPIModel` in `internal/elasticsearch/index/template/expand.go`, after the existing `Version` block, add: when `m.AllowAutoCreate` is non-null and non-unknown, set `out.AllowAutoCreate = m.AllowAutoCreate.ValueBoolPointer()`.

## 7. Flatten

- [ ] 7.1 In `(m *Model).fromAPIModel` in `internal/elasticsearch/index/template/flatten.go`, after the existing `Priority`/`Version` block, add: `m.AllowAutoCreate = boolFromBoolPtr(in.AllowAutoCreate)` where `boolFromBoolPtr` mirrors the existing `int64FromInt64Ptr` helper: returns `types.BoolNull()` for nil, `types.BoolValue(*p)` otherwise. Add the helper function to `flatten.go`.

## 8. Acceptance tests

- [ ] 8.1 In the index template acceptance test file (`internal/elasticsearch/index/template/acc_test.go` or equivalent), add a test `TestAccResourceIndexTemplateAllowAutoCreate` that:
  - Creates a template with `allow_auto_create = true`, asserts state `allow_auto_create = true`.
  - Updates to `allow_auto_create = false`, asserts state `allow_auto_create = false`.
  - Removes the attribute (sets null), verifies no diff after apply (omitted from request).
  - Runs an import step and verifies `allow_auto_create` is populated from the API response.
- [ ] 8.2 Verify that existing acceptance tests continue to pass (no unintended drift from adding the attribute with a null default).

## 9. Documentation

- [ ] 9.1 Regenerate provider docs (`make docs` or equivalent) so `allow_auto_create` appears in the generated resource and data source documentation pages.

## 10. OpenSpec

- [ ] 10.1 Keep delta spec `openspec/changes/index-template-allow-auto-create/specs/elasticsearch-index-template/spec.md` aligned with the final implementation.
- [ ] 10.2 After merge: sync delta into `openspec/specs/elasticsearch-index-template/spec.md` or archive the change per project workflow; run `make check-openspec`.
