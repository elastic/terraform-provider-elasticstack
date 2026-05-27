## Why

The `elasticstack_kibana_slo` resource already supports ES Query DSL (including `regexp`, `wildcard`, and `bool` queries) through the `filter_kql.filters[*].query` attribute introduced in commit `5a4eb5c1` (2026-04-28). However, this capability is not documented anywhere — the resource docs only show KQL string examples, and the schema description for `query` does not mention ES Query DSL.

The original reporter (issue #1049) wanted to filter SLO data with a `regexp` query on `httpRequest.referer` to match CDN requests from PR-preview environments — a use case that cannot be expressed in plain KQL but is fully expressible today using `jsonencode({regexp: {...}})` inside `filter_kql.filters[*].query`. Without documentation, users cannot discover this capability.

## What Changes

- Update `docs/resources/kibana_slo.md` to add a worked example of a `regexp` query inside `kql_custom_indicator.filter_kql.filters[*].query` using `jsonencode`.
- Add or update an entry in `examples/resources/elasticstack_kibana_slo/resource.tf` demonstrating the same ES Query DSL filter pattern.
- Update the schema description for `attrQuery` inside `kqlWithFiltersObjectSchema` in `internal/kibana/slo/schema.go` to explicitly state that the field accepts any ES Query DSL JSON object (not just KQL-compatible queries).

No code logic changes are required. The API round-trip is already implemented and the schema accepts arbitrary normalized JSON.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-slo`: add documentation and a schema description clarifying that `filter_kql.filters[*].query` (and equivalents on `good_kql` / `total_kql`) accepts any ES Query DSL JSON, including `regexp`, `wildcard`, `bool`, and similar query types.

## Impact

- Affected files are limited to `docs/resources/kibana_slo.md`, `examples/resources/elasticstack_kibana_slo/resource.tf`, and the `attrQuery` description string in `internal/kibana/slo/schema.go`.
- No schema version changes. No model or test logic changes.
- The change is purely additive and carries zero runtime risk.
