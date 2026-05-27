# Delta Spec: `elasticstack_kibana_slo` — ES Query DSL documentation for `kql_custom_indicator` filters

Base spec: `openspec/specs/kibana-slo/spec.md`
Last requirement in base spec: REQ-040
This delta introduces: REQ-041

---

This delta documents the existing capability — implemented on `main` in commit `5a4eb5c1` (2026-04-28) — to use arbitrary ES Query DSL expressions inside the `filter_kql.filters[*].query`, `good_kql.filters[*].query`, and `total_kql.filters[*].query` attributes of `kql_custom_indicator`. The `query` attribute is typed `jsontypes.NormalizedType{}`, which accepts any valid JSON object without schema-level constraint, making it suitable for `regexp`, `wildcard`, `bool`, `range`, and other ES Query DSL query types.

No code logic is changed by this delta. The work is limited to updating the schema description for `query` in `kqlWithFiltersObjectSchema()` and adding a documentation example.

## ADDED Requirements

### Requirement: Documentation — ES Query DSL filter example in `kql_custom_indicator` (REQ-041)

The provider documentation and examples for `elasticstack_kibana_slo` SHALL include a worked example demonstrating how to use an ES Query DSL `regexp` query as the value of `filter_kql.filters[*].query` within a `kql_custom_indicator` block. The example SHALL use the `jsonencode({...})` helper so practitioners can construct the required JSON string inline.

The schema attribute description for `query` inside the `filters` list (used by `filter_kql`, `good_kql`, and `total_kql`) SHALL explicitly state that the field accepts an ES Query DSL JSON object (including `regexp`, `wildcard`, `prefix`, `range`, and `bool` combinators; support is subject to the Kibana SLO API / stack version) and SHALL reference the `jsonencode({...})` usage pattern.

#### Scenario: Regexp filter example is present in docs

- GIVEN the `docs/resources/kibana_slo.md` documentation page
- WHEN a user looks for guidance on using regexp or advanced ES Query DSL filters in an SLO
- THEN the documentation SHALL include an HCL example with `filter_kql.filters` and a `jsonencode({regexp: {...}})` query value

#### Scenario: ES Query DSL schema description is present on the `query` attribute

- GIVEN the Terraform schema for `kql_custom_indicator` in the resource documentation
- WHEN a user reads the attribute description for `query` inside `filter_kql.filters`, `good_kql.filters`, or `total_kql.filters`
- THEN the description SHALL state that the field accepts any ES Query DSL JSON object and SHALL reference the `jsonencode({...})` pattern

#### Scenario: Example uses the reporter's regexp pattern

- GIVEN the reporter's use case from issue #1049 — filtering `httpRequest.referer` with a `regexp` matching PR-environment CDN URLs
- WHEN the documentation example is rendered
- THEN the `regexp` query block SHALL include `case_insensitive`, `flags`, and `value` keys, matching the shape from the issue
