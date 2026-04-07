# Tasks: Require `searchable_snapshot` in the ILM `frozen` phase

## 1. Spec

- [x] 1.1 Keep the delta spec aligned with proposal.md and design.md
- [x] 1.2 On completion, sync the delta into the canonical spec or archive the change

## 2. Schema and validation

- [x] 2.1 Update `internal/elasticsearch/index/ilm/schema.go` so `frozen.searchable_snapshot` is required when the `frozen` phase is configured
- [x] 2.2 Ensure ILM validation rejects a `frozen` phase that omits `searchable_snapshot` before any Elasticsearch API call
- [x] 2.3 Preserve the existing requirement that `searchable_snapshot.snapshot_repository` is required when the action block is present

## 3. Documentation

- [x] 3.1 Regenerate `docs/resources/elasticsearch_index_lifecycle.md`
- [x] 3.2 Confirm the generated docs describe `frozen.searchable_snapshot` as required within the `frozen` phase

## 4. Testing

- [x] 4.1 Keep or update acceptance coverage for a valid `frozen` phase with `searchable_snapshot`
- [x] 4.2 Add validation-focused test coverage for `frozen {}` without `searchable_snapshot`
- [x] 4.3 Run the relevant ILM tests and any required OpenSpec validation
