## Why

`elasticstack_elasticsearch_cluster_settings` emits a hard validation error — "No cluster settings
configured" — when `persistent` and/or `transient` blocks are populated via `dynamic` blocks whose
`for_each` is driven by local values. This is a regression introduced in the SDKv2 → Plugin
Framework migration (v0.15.0) that added the `ValidateConfig` hook. Terraform calls
`ValidateResourceConfig` before local values are evaluated; at that point the block objects are
**unknown**, not null, but the current `categoryBlockEmpty` helper treats unknown identically to
null and therefore emits a false-positive error.

The fix aligns the helper with the Plugin Framework convention: unknown means "we don't have
enough information yet to validate" — it does NOT mean "this value is absent". Every standard
validator in `terraform-plugin-framework-validators` and the project's own
`settingNameUniqueValidator` already follow this convention by returning early on unknown.

## What Changes

- Fix `categoryBlockEmpty` in `internal/elasticsearch/cluster/settings/resource.go` to return
  `false` (not empty) when the outer block is unknown, rather than `true`. Apply the same fix for
  an unknown inner `setting` set.
- Add three unit tests to `helpers_test.go` (and the necessary export hook to `export_test.go` if
  needed) to lock in correct behavior:
  - Both blocks unknown → no error
  - One block null, other unknown → no error
  - Direct `categoryBlockEmpty` helper test: unknown block → returns `false`

## Capabilities

### Modified Capabilities

- `elasticsearch-cluster-settings`: `ValidateConfig` no longer emits false
  errors when `persistent`/`transient` blocks are populated by unevaluated `dynamic for_each`
  expressions.

## Impact

- Single function change in `internal/elasticsearch/cluster/settings/resource.go` (3 lines).
- Three new unit tests in `internal/elasticsearch/cluster/settings/helpers_test.go`.
- No changes to API interaction logic, state upgrade paths, schema, or acceptance tests.
- All existing unit tests continue to pass unchanged.
