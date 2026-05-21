## Why

`elasticstack_elasticsearch_component_template` fails with "Provider produced inconsistent result
after apply" immediately after `terraform apply` succeeds. Three distinct errors are observed:

1. **`.version` drift** — planned `null`, state after apply `cty.NumberIntVal(0)`.
2. **`.template[0].mappings` drift** — planned `null` (no mappings block), state `cty.StringVal("")`.
3. **`.template[0].settings` format change** — plan stored `{"number_of_shards":"3"}`, read-back
   stored `{"index":{"number_of_shards":"3"}}` (Elasticsearch wraps short-form keys on round-trip).

Errors 1 and 3 were resolved in v0.15.0 by the Plugin Framework migration (commit `d73b1170`). Error
2 (and a related `null vs "{}"` inconsistency) is still reachable: some Elasticsearch versions return
`"mappings": {}` or `"settings": {}` in `GET /_component_template` for templates that have no
mappings or settings defined. The current `flattenTemplateBlock` tests `!= nil` to decide whether to
emit a value, but Go's JSON decoder sets a field to a non-nil empty map when the API returns `{}`.
The framework then sees `null` (plan) vs `"{}"` (state) and fails the post-apply consistency check.

## What Changes

Harden `flattenTemplateBlock` in
`internal/elasticsearch/index/componenttemplate/flatten.go` to treat empty maps (`len == 0`) the
same as nil — both should produce a null Terraform value. This closes the `null vs "{}"` gap with a
two-line targeted change.

Add a no-drift acceptance test step to `TestAccResourceComponentTemplate` that re-applies the
existing create config and asserts an empty plan, confirming neither `mappings` nor `settings`
produce spurious post-apply drift. Add a dedicated test config (`issue-609`) mirroring the exact
original reporter scenario: alias present, short-form `number_of_shards = "3"` at the top level of
the `settings` object (not nested under `index = {}`), and no `mappings` block.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `elasticsearch-index-component-template`: extend REQ-022 (read state mapping for `template.mappings`
  and `template.settings`) to explicitly cover the empty-object (`{}`) Elasticsearch response —
  treating it as equivalent to absent (`null`).

## Impact

- `internal/elasticsearch/index/componenttemplate/flatten.go` — `flattenTemplateBlock`: change
  `t.Mappings != nil` to `len(t.Mappings) > 0` and `t.Settings != nil` to `len(t.Settings) > 0`.
- `internal/elasticsearch/index/componenttemplate/acc_test.go` — add no-drift `PlanOnly` step to
  `TestAccResourceComponentTemplate`; add a new `TestAccResourceComponentTemplateIssue609NoDrift`
  test using the `apply` config directory.
- `internal/elasticsearch/index/componenttemplate/testdata/TestAccResourceComponentTemplateIssue609NoDrift/apply/main.tf` —
  new test config mirroring the original issue: alias, top-level `number_of_shards = "3"` in
  settings, no mappings block.
