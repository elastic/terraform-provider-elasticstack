## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate add-kibana-tag-support --type change` (or `make check-openspec` after sync).
- [ ] 1.2 Confirm name and color validation rules during implementation (open questions in `design.md`); update delta spec with validators if confirmed.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/kibana-tag/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Add `internal/models/tag.go` with `Tag` struct: `ID`, `Name`, `Color`, `Description`, `SpaceID`, `CreatedAt`, `UpdatedAt` fields mapped from `KibanaHTTPAPIsKbnTagsAttributes` + `KibanaHTTPAPIsKbnAsCodeMeta`.
- [ ] 2.2 Add `internal/clients/kibanaoapi/tag.go` wrapping the kbapi `Client` methods: `GetTag(ctx, spaceID, id)`, `CreateTag(ctx, spaceID, req)`, `UpsertTag(ctx, spaceID, id, req)`, `DeleteTag(ctx, spaceID, id)`, `ListTags(ctx, spaceID, params)`. Handle space-aware routing via the existing KibanaAPI client space mechanism.
- [ ] 2.3 Create `internal/kibana/tag/` package with the following files:
  - `resource.go` — Plugin Framework resource registration and CRUD methods
  - `schema.go` — resource schema with all attributes (name, tag_id, color, description, space_id, id, created_at, updated_at)
  - `models.go` — `tagModel` Go struct for Plugin Framework state; `toAPIModel()` and `fromAPIModel()` helpers
  - `create.go` — branching create logic (POST when tag_id absent, GET+PUT when tag_id present)
  - `read.go` — GET + managed-tag guard
  - `update.go` — PUT + managed-tag guard
  - `delete.go` — DELETE + managed-tag guard
  - `guard.go` — `managedTagDiagnostic()` and `checkManagedTag()` helpers (pattern from `osquery_pack/guard.go`)
  - `import.go` — ImportState parsing `"<space_id>/<id>"` with managed guard on resulting read
- [ ] 2.4 Create data source files under `internal/kibana/tag/`:
  - `datasource.go` — Plugin Framework data source registration and Read method
  - `datasource_schema.go` — data source schema (query, space_id, tags list)
  - `datasource_models.go` — `tagsDataSourceModel` and `tagItemModel` structs
  - `datasource_read.go` — auto-paginating list via `ListTags`; returns empty list on no results
- [ ] 2.5 Add Kibana ≥ 9.5.0 version gate in the resource Create/Read/Update/Delete and data source Read entry points, using the existing version-check helper. Return a clear `ErrorDiagnostic` on mismatch.
- [ ] 2.6 Register `elasticstack_kibana_tag` resource and `elasticstack_kibana_tags` data source in the provider registration (wherever other kibana resources and data sources are registered).
- [ ] 2.7 Add embedded schema attribute descriptions (in `schema.go` or a `descriptions/` sub-package consistent with the rest of the `internal/kibana/tag/` package style).
- [ ] 2.8 Generate or update provider documentation for the new resource and data source (run `make generate-docs` or equivalent).

## 3. Testing

- [ ] 3.1 Add acceptance test `TestAccResourceKibanaTag_basic` in `internal/kibana/tag/acc_test.go`: create a tag with `name` and `color`; assert state; update `name`; assert updated state; destroy and assert gone.
- [ ] 3.2 Add acceptance test `TestAccResourceKibanaTag_noColor`: create a tag without `color`; assert `color` is Computed and non-empty; update name; assert `color` is unchanged (no diff).
- [ ] 3.3 Add acceptance test `TestAccResourceKibanaTag_withTagID`: create a tag with an explicit `tag_id`; assert `id = "<space>/<tag_id>"`; verify import works with `"<space>/<tag_id>"` format.
- [ ] 3.4 Add acceptance test `TestAccResourceKibanaTag_import`: create via Terraform; run `terraform import`; assert state is consistent after import.
- [ ] 3.5 Add acceptance test `TestAccDataSourceKibanaTags_basic`: create two tags then query data source; assert both appear in `tags` list.
- [ ] 3.6 Add acceptance test `TestAccDataSourceKibanaTags_query`: create tags with distinct names; query with `query = "<name>"` filter; assert only matching tags returned.
- [ ] 3.7 All acceptance tests MUST be skipped when the connected Kibana version is below 9.5.0 (use existing skip-version helper or add one consistent with other resources).
- [ ] 3.8 Add unit tests for `toAPIModel()` (nil description, nil color), `fromAPIModel()` (populated and sparse responses), and the create-branching logic (POST path vs PUT path).
- [ ] 3.9 Add unit test for managed-tag guard: simulate `meta.managed = true` response; assert `checkManagedTag` returns a non-nil error diagnostic.
