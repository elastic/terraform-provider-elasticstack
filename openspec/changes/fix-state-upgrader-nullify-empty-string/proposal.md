# Proposal: Fix state upgrader — nullify empty-string JSON fields

## Problem

Upgrading `elasticstack_elasticsearch_component_template` from provider ≤0.14.x (Plugin SDK v2) to
≥0.15.0 (Plugin Framework) causes a fatal `terraform plan` error for settings-only templates (no
`mappings` block):

```
Error: Semantic Equality Check Error
unexpected end of JSON input
```

**Root cause:** Plugin SDK v2 serializes unset optional strings as `""` rather than `null`. State for
a settings-only template therefore contains `"mappings": ""`. The state upgrader
`migrateComponentTemplateStateV0ToV1` calls `stateutil.EnsureMapKeys` (which initialises absent keys
to `nil`) but never calls `stateutil.NullifyEmptyString`, so `"mappings": ""` passes through
unchanged. During `terraform plan`, `MappingsValue.StringSemanticEquals` calls
`json.Unmarshal([]byte(""), &v)`, which returns `unexpected end of JSON input` and aborts the plan.

The same omission exists in `migrateIndexTemplateStateV0ToV1`, which likely caused the
"Failed to decode resource from state: missing expected {" warnings reported on
`elasticstack_elasticsearch_index_template` resources after the same upgrade.

## Recommendation

Fix the root cause in both state upgraders by adding `stateutil.NullifyEmptyString` calls
immediately after `stateutil.EnsureMapKeys` for all JSON-string attributes inside `template`
(`mappings`, `settings`). The identical helper is already applied by the transform resource upgrader
for exactly this pattern and is proven to work.

Additionally, check whether `metadata` (a top-level JSON string attribute on both resources) is
stored as `""` by the SDK when unset, and nullify it if so.

## Scope

- `internal/elasticsearch/index/componenttemplate/state_upgrade.go` — add `NullifyEmptyString` for
  `mappings` and `settings` inside the `template` block; nullify top-level `metadata` if it arrives
  as `""`.
- `internal/elasticsearch/index/template/state_upgrade.go` — same fix for `mappings` and `settings`
  inside the `template` block; nullify top-level `metadata` if it arrives as `""`.
- `internal/elasticsearch/index/componenttemplate/state_upgrade_test.go` — add unit test cases for
  the settings-only (empty-string `mappings`) path and the `metadata: ""` path.
- `internal/elasticsearch/index/template/state_upgrade_test.go` — add the same unit test cases for
  index template.
- `internal/elasticsearch/index/componenttemplate/acc_from_sdk_test.go` — add a settings-only
  variant to `TestAccResourceComponentTemplateFromSDK` (additional test step or separate test) that
  exercises the SDK → PF upgrade with no `mappings` block.
- `internal/elasticsearch/index/template/acc_from_sdk_test.go` — add a settings-only variant to
  `TestAccResourceIndexTemplateFromSDK` confirming the fix is symmetric.

## Out of scope

- Changes to `MappingsValue.StringSemanticEquals` or any other semantic equality function.
- The perpetual diff caused by flat-key vs. nested settings (`"index.lifecycle.name"` vs.
  `{"index":{"lifecycle":{"name":"..."}}}`). That is a separate UX issue.
- Any Elasticsearch API, schema, or documentation changes.
