## Context

After commit `5a4eb5c1` (2026-04-28), the `elasticstack_kibana_slo` resource accepts object-form KQL for `kql_custom_indicator` through three additive attributes: `filter_kql`, `good_kql`, and `total_kql`. Each attribute exposes a `kql_query` string and a `filters` list. Each filter item has a `query` field typed as `jsontypes.NormalizedType{}`, which accepts any JSON object.

Because `jsontypes.NormalizedType{}` is schema-free, it can hold any ES Query DSL expression — including `regexp`, `wildcard`, `prefix`, `range`, and arbitrarily nested `bool` combinators — without any provider-side code changes. This capability was not mentioned in the docs, the schema descriptions, or any example file.

The implementation research for issue #1049 confirmed this and recommended documenting the existing capability (Approach A).

## Goals / Non-Goals

**Goals:**
- Document the `jsonencode({regexp: {...}})` pattern for `filter_kql.filters[*].query` clearly enough that a new user can copy-paste a working example.
- Clarify in the schema description that `query` is an arbitrary ES Query DSL JSON object, not just a KQL-compatible filter.
- Update the generated `docs/resources/kibana_slo.md` after the description change.

**Non-Goals:**
- Code changes to model serialization, schema structure, state upgraders, or tests.
- Adding `meta.negate` or other `SLOsFilterMeta` fields (Approach B in the research; a separate follow-on).
- Adding ES Query DSL filter support to `metric_custom_indicator`, `histogram_custom_indicator`, or `apm_*` indicators — those indicator types use a plain `*string` KQL field at the Kibana API layer.

## Decisions

### 1. Use `jsonencode` in the example, not raw JSON string

The `query` attribute is `jsontypes.NormalizedType{}`, so practitioners must provide a valid JSON string. The idiomatic Terraform approach is `jsonencode({...})`. The example should show this pattern explicitly so readers understand why the value is wrapped.

### 2. Show the `regexp` query from the issue verbatim

The reporter's exact query (`regexp` on `httpRequest.referer` with `case_insensitive`, `flags`, and `value`) should be the primary example because it directly validates the reported use case. Additional context (e.g., noting that `wildcard` or `bool` queries work the same way) can appear as a comment.

### 3. Schema description update is low-risk

The `attrQuery` description change touches a single string constant in `schema.go`. No logic is affected, and `terraform-plugin-docs` will regenerate the attribute's description block in `kibana_slo.md` on the next docs run. The implementer should run `make generate-docs` (or equivalent) after updating the description.

### 4. Examples file update alongside docs

`examples/resources/elasticstack_kibana_slo/resource.tf` is the source used by `terraform-plugin-docs` for the `## Example Usage` section. Adding the `regexp` example there keeps the generated docs accurate and also helps users browsing the examples directory directly.

## Risks / Trade-offs

- **API may validate query content** — The Kibana SLO backend types `SLOsFilter.Query` as `map[string]interface{}` in the OpenAPI spec (no constraint). Whether Kibana enforces KQL compatibility at runtime is an open question (see Open Questions). The documentation should include a note that the backend is expected to accept any ES Query DSL JSON, but this has not been verified with an integration test against a live cluster.
- **Docs regeneration required** — If the description is updated in `schema.go` but `make generate-docs` is not run, the markdown file will be stale. This is standard provider procedure but worth noting in the task.

## Open Questions

- Is `filter_kql.filters` included in a released version, or only on `main`? If the next release is not imminent, a comment on issue #1049 pointing to the unreleased feature would help the reporter.
- Does the Kibana SLO backend actually accept a raw `regexp` query in `filter.query`, or does it validate KQL compatibility? The OpenAPI spec types `SLOsFilter.Query` as `map[string]interface{}` (no constraint), but an integration test would confirm.
- Should the `meta.negate` gap be filed as a follow-on issue before this one is closed?
