## Why

Users who manage a mix of explicit (Terraform-owned) and dynamic (auto-generated) field mappings on the same index experience persistent plan drift when running `terraform plan` across different workspaces. Each workspace's state accumulates different sets of dynamic fields over time. The existing `elasticstack_elasticsearch_index` resource has semantic-equality helpers that partially mitigate this, but the merge-into-state approach cannot eliminate drift when the same index is managed by multiple workspaces or not owned by Terraform at all (e.g. created by ILM policy, data stream rollover, index template, or external tooling).

The concrete user need is a separate `elasticstack_elasticsearch_index_mappings` resource that:
1. Targets an **existing** index (or one managed by another resource).
2. Manages a specific subset of explicit top-level mapping definitions via `PUT /{index}/_mapping`.
3. Only tracks what the user declares in state — dynamic extras returned by the API are silently dropped on read — so drift is eliminated by design.

## What Changes

- **New resource `elasticstack_elasticsearch_index_mappings`**: A standalone resource in `internal/elasticsearch/index/indexmappings/` following the same sub-resource pattern used by `elasticstack_elasticsearch_index_alias` and `elasticstack_elasticsearch_data_stream_lifecycle`.
  - Create / Update → `PUT /{index}/_mapping`; errors immediately if the target index does not exist.
  - Read → retrieves index metadata via `GetIndex`; state is repopulated with only the user-declared subset. The intersection mask comes from the **previously stored state** (not the current configuration), which is the data available to the read callback in the `entitycore` envelope. For each top-level key present in prior state, the value is retained; within `properties`, only field names present in the prior state's `properties` tree are retained recursively. Dynamic extras are silently dropped. On import (no prior mask), the full API response is stored initially.
  - Delete → **no-op**: field mappings cannot be removed without a full reindex; the resource is removed from state only with no API call.
  - The `mappings` attribute accepts a JSON string (using the existing `MappingsType` custom type from `internal/elasticsearch/index/mappings_value.go`) that covers all top-level mapping parameters (`properties`, `dynamic`, `_source`, `dynamic_templates`, `runtime`, etc.).
  - `index` attribute is required and forces resource replacement when changed.
  - ID format: `{cluster_uuid}/{index_name}`.
  - Supports `elasticsearch_connection` block (injected by the provider scaffold).

- **No changes to `elasticstack_elasticsearch_index`**: Existing resource semantics are unaffected.

## Capabilities

### New Capabilities

- `elasticstack-elasticsearch-index-mappings`: New resource managing a user-declared subset of Elasticsearch index mappings without drift, independent of index lifecycle ownership.

### Modified Capabilities

None.

## Design decisions baked in

The following open questions from the research comment have been resolved by the issue author (@tobio):

| Question | Decision |
|---|---|
| Destroy semantics | Remove from state only (no-op). |
| Data stream support | Out of scope for this change. |
| Index existence at create | Error if the target index does not exist. |
| Scope beyond `properties` | All top-level mapping keys accepted. |
| Conflict detection | Left to the user; no provider-level warning. |

## Impact

- New package `internal/elasticsearch/index/indexmappings/` (`resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`).
- Provider registration in `provider.go` to expose the new resource type.
- Acceptance tests in `internal/elasticsearch/index/indexmappings/acc_test.go` and corresponding `testdata/` fixtures.
- Documentation page generated from schema (no manual edits required).
- No impact on existing resources, data sources, or CI workflows.
