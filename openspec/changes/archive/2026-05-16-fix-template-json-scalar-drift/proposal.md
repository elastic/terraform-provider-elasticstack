## Why

The 0.15.x Plugin Framework migration regressed semantic handling of JSON scalars in Elasticsearch template JSON attributes. `elasticstack_elasticsearch_component_template` now raises post-apply consistency errors when Elasticsearch echoes scalar values as strings, such as `true` becoming `"true"` in `template.mappings` (#2987) and `null` becoming `"null"` in `template.settings` (#2988).

## What Changes

- Update shared template JSON custom-type behavior so Elasticsearch stringified scalar echoes are treated as semantically equal to practitioner-authored scalar JSON values where appropriate.
- Extend mapping semantic equality to tolerate scalar-vs-string scalar equivalence at leaf values while preserving existing structural comparison and non-drifting superset behavior.
- Update index settings canonicalization so JSON `null` is normalized consistently with Elasticsearch's `"null"` string echo.
- Add unit coverage for shared custom types and focused acceptance coverage for component template mappings/settings regressions.
- Update OpenSpec requirements for component templates and index templates to explicitly describe scalar-string semantic equivalence for mappings/settings refresh behavior.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-index-component-template`: `template.mappings` and `template.settings` refresh semantics will explicitly tolerate Elasticsearch stringified scalar echoes that are semantically equivalent to practitioner-authored scalar JSON values.
- `elasticsearch-index-template`: shared `template.mappings` and `template.settings` custom-type semantics will explicitly tolerate Elasticsearch stringified scalar echoes that are semantically equivalent to practitioner-authored scalar JSON values.

## Impact

- `internal/elasticsearch/index/mappings_value.go` and `internal/elasticsearch/index/mappings_value_test.go`
- `internal/utils/customtypes/index_settings_value.go` and related tests
- `internal/elasticsearch/index/componenttemplate/acc_test.go` and new/updated acceptance testdata
- `openspec/specs/elasticsearch-index-component-template/spec.md`
- `openspec/specs/elasticsearch-index-template/spec.md`
