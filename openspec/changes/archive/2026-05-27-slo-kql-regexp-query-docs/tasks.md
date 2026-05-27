## 1. Schema description

- [x] 1.1 In `internal/kibana/slo/schema.go`, inside `kqlWithFiltersObjectSchema()`, update the `attrQuery` description to explicitly state that the `query` field accepts any ES Query DSL JSON object — including `regexp`, `wildcard`, `prefix`, `range`, and `bool` combinators — and is not limited to KQL-compatible filter shapes. Suggested new description: `"Filter query as a JSON-encoded ES Query DSL object. Accepts any valid Elasticsearch Query DSL expression (regexp, wildcard, bool, range, etc.). Use jsonencode({...}) to construct the value in Terraform configuration."`

## 2. Example and generated docs

- [x] 2.1 Add a `kql_custom_indicator` configuration block to `examples/resources/elasticstack_kibana_slo/resource.tf` that demonstrates the `filter_kql.filters[*].query` attribute with a `jsonencode({regexp: {...}})` value, using the reporter's exact query pattern (`httpRequest.referer` with `case_insensitive`, `flags`, and `value` keys). Include a comment noting that `good_kql` and `total_kql` support the same `filters` pattern.

- [x] 2.2 Run `make generate-docs` (or the equivalent terraform-plugin-docs command) to regenerate `docs/resources/kibana_slo.md` with the updated schema description from task 1.1 and the new example from task 2.1.

- [x] 2.3 Review the regenerated `docs/resources/kibana_slo.md` to confirm the new `regexp` example renders correctly and the `query` attribute description reflects ES Query DSL support.
