## Context

Canonical requirements live in [`openspec/specs/elasticsearch-index-template/spec.md`](../../specs/elasticsearch-index-template/spec.md). The implementation lives in [`internal/elasticsearch/index/template/`](../../../internal/elasticsearch/index/template/).

The `go-elasticsearch` v8 typed client already models `AllowAutoCreate *bool` on `estypes.IndexTemplate` (field `AllowAutoCreate`). The internal `models.IndexTemplate` struct in `internal/models/models.go` does not yet include this field; the Terraform `Model` struct and schema likewise do not expose it.

The expand/flatten pattern for optional nullable bool fields (e.g. `priority`, `version`) sets the field only when non-null during expand and reads back the pointer value during flatten. `allow_auto_create` follows the same pattern.

## Goals / Non-Goals

**Goals:**

- Expose `allow_auto_create` as an optional top-level attribute on the resource and computed on the data source.
- Round-trip correctly: set → expand → PUT → GET → flatten → state matches configured value.
- No state migration overhead: the field defaults to null in existing state, which is backward-compatible.

**Non-goals:**

- Version gating: not required; `allow_auto_create` is available from ES 7.11, below provider minimum 7.17.
- Any change to cluster-level `action.auto_create_index` or `elasticstack_elasticsearch_cluster_settings`.
- Exposing other unexposed `IndexTemplate` fields (e.g. `deprecated`).

## Decisions

- **Null vs false**: When `allow_auto_create` is null (not configured), the field is omitted from the API request (`omitempty`), matching the behavior of `priority` and `version`. When explicitly set to `false`, the field **is sent** to Elasticsearch with value `false` — this is intentional because the provider minimum is ES 7.17 (above 7.11 where the field was introduced), so sending explicit `false` is safe and semantically meaningful (overrides a cluster `action.auto_create_index = true`). This differs from `data_stream.allow_custom_routing`, which gates sending on `true` only for pre-8.x compatibility; no such constraint applies here.
- **Data source**: The field is added as `Computed: true` so the data source read path can populate it from the API response.
- **Descriptions**: A new constant `descAllowAutoCreate` is added to `descriptions.go`.
- **No `Default`**: Unlike `data_stream.hidden` and `allow_custom_routing`, `allow_auto_create` carries no schema-level default. The field is left null when not configured, consistent with `priority` and `version`.

## Risks / Trade-offs

- **Minimal risk**: The change is purely additive. No existing behavior is altered. Adding a nullable optional attribute with `omitempty` JSON serialization does not affect existing payloads.
- **Acceptance test scope**: Tests need to cover the `false` path explicitly because it is the less obvious case (explicitly sent to ES, not omitted).

## Open Questions

- Should `allow_auto_create = false` be sent explicitly to ES (to override a cluster `action.auto_create_index = true`) or treated as "omit the field" when false (same as null)? **Decision**: send explicit `false` — the provider minimum is ES 7.17 (above 7.11 where the field landed), so explicit `false` is safe and semantically useful.
- Should the data source expose `allow_auto_create`? **Decision**: yes — parity with the resource read path, and `GetIndexTemplate` already returns it via `estypes.IndexTemplate.AllowAutoCreate`.
- Does any acceptance test fixture need updating for the state upgrade path? **Decision**: no — the purely additive null-default is backward-compatible without a state version bump or upgrader.
