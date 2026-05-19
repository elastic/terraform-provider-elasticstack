## Why

The `elasticstack_elasticsearch_index_template` resource does not expose the `allow_auto_create` field supported by the Elasticsearch [PUT index template API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-indices-put-index-template). This top-level boolean allows operators to control per-template whether matching indices may be auto-created, overriding the cluster-level `action.auto_create_index` setting — a capability that is currently inaccessible through the provider.

## What Changes

- Add `allow_auto_create` as an **optional** top-level `BoolAttribute` on the `elasticstack_elasticsearch_index_template` resource schema at the same level as `priority`, `version`, and `composed_of`.
- Add `allow_auto_create` as a **computed** `BoolAttribute` on the `elasticstack_elasticsearch_index_template` data source schema.
- Add `AllowAutoCreate *bool` with `json:"allow_auto_create,omitempty"` to `internal/models/models.go` (`IndexTemplate` struct).
- Add `AllowAutoCreate types.Bool` field to `internal/elasticsearch/index/template/models.go` (`Model` struct).
- Wire expand: in `toAPIModel`, set `out.AllowAutoCreate` from the model when non-null.
- Wire flatten: in `fromAPIModel`, copy `in.AllowAutoCreate` to `m.AllowAutoCreate`.
- Add description constant `descAllowAutoCreate` in `descriptions.go`.
- Add acceptance test coverage: create/update with `allow_auto_create = true`, import round-trip, and verify null default (no drift when omitted).

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-index-template`**: Schema and behavior requirements updated to include the `allow_auto_create` attribute on the resource (Optional) and data source (Computed).

## Impact

- **Users**: Can now set `allow_auto_create = true` (or `false`) on an index template to override the cluster-level auto-create setting on a per-template basis. Existing configurations are unaffected — omitting the attribute leaves its value null (not sent to Elasticsearch), preserving prior behavior.
- **Code**: `internal/models/models.go`, `internal/elasticsearch/index/template/models.go`, `schema.go`, `data_source_schema.go`, `expand.go`, `flatten.go`, `descriptions.go`, and the acceptance test file.
- **State**: Additive null-default attribute. No state migration required; existing state reads null for this field, consistent with "not set."
- **Elasticsearch version**: `allow_auto_create` on index templates is available since ES 7.11, below the provider's minimum (7.17+). No version gate is needed.
