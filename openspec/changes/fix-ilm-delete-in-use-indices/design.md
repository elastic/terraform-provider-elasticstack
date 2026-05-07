## Context

The `elasticstack_elasticsearch_index_lifecycle` resource Delete handler (`internal/elasticsearch/index/ilm/delete.go`) calls the Elasticsearch Delete Lifecycle API directly. Elasticsearch rejects the call with `illegal_argument_exception` when one or more indices still reference the policy in `index.lifecycle.name`.

This failure is common in Fleet-managed data stream workflows:

```
Fleet integration install ──► index template + component template
                                    │
                                    ▼
                         data stream created (ES auto-creates backing index)
                                    │
                                    ▼
         ILM attachment resource adds lifecycle.name to @custom component template
                                    │
                                    ▼
                         backing index inherits lifecycle.name
                                    │
                                    ▼
Fleet uninstall succeeds but backing index remains ◄── data intentionally preserved
                                    │
                                    ▼
ILM policy delete fails ◄── ES: "in use by one or more indices"
```

The error is surfaced by the provider as a Terraform diagnostic with the verbatim ES error, forcing the user to manually null `index.lifecycle.name` on every matching index.

## Goals / Non-Goals

**Goals:**
- Make ILM policy Delete succeed when indices happen to reference the policy.
- Preserve all index data — only remove the ILM reference, not the index.
- Keep the change localized to the ILM resource and ES client helpers.

**Non-Goals:**
- Do **not** delete indices, data streams, or Fleet-managed assets. Data deletion is out of scope.
- Do **not** change Fleet integration uninstall behavior.
- Do **not** introduce a new schema attribute (e.g., `force`). The fix is transparent.

## Decisions

### Decision: Scan-and-null approach

Before calling `DELETE /_ilm/policy/{name}`, the Delete handler will:

1. Fetch all index settings for the key `index.lifecycle.name` via:
   `GET /_all/_settings/index.lifecycle.name?flat_settings=true`
2. Filter to indices whose `index.lifecycle.name` equals the policy name being deleted.
3. If any matches exist, issue:
   `PUT /{indices_comma_separated}/_settings {"index.lifecycle.name": null}`
4. Then proceed with the existing `DELETE /_ilm/policy/{name}` call.

**Rationale**: This is the smallest intervention that solves the problem. It mirrors what users do manually. Using `_all` with `flat_settings` is a single lightweight call.

**Alternative considered**: Maintain a reverse-lookup map of which templates reference this policy and clear ILM from templates first, then rollover indices. Rejected because backing indices can outlive template changes — we must target the indices directly anyway.

### Decision: Batch the settings update per matched policy

If indices match, they are collected into one comma-separated target and cleared in a single `PUT /{indices}/_settings` call rather than N sequential calls.

**Rationale**: Reduces ES API pressure. Terraform destroy may already be deleting many resources; adding N extra round-trips per ILM policy is wasteful.

### Decision: No retry or rollback on partial failure

If the settings-clear call partially succeeds (some indices cleared, some still referencing the policy), the subsequent `DELETE /_ilm/policy/{name}` will still fail with an updated "in use by" message. We surface that error.

**Rationale**: The error message tells the user exactly which indices remain. Adding retry logic adds complexity for an edge case that usually indicates a transient ES state (index closed, relocating, etc.) or a race with another system.

### Decision: ES client helper functions

Two new functions in `internal/clients/elasticsearch/index.go`:

- `GetIndicesWithILMPolicy(ctx, client, policyName) ([]string, diag.Diagnostics)`
  - Wraps `GET /_all/_settings/index.lifecycle.name?flat_settings=true`
  - Returns just the index names that reference the given policy

- `ClearILMPolicyFromIndices(ctx, client, indices []string) diag.Diagnostics`
  - Wraps `PUT /{indices}/_settings {"index.lifecycle.name": null}`

These are general-purpose enough that they may be reused by other resources later (e.g., index template ILM attachment cleanup).

**Rationale**: Keeps the ILM resource handler thin and places ES-API-specific logic in the client layer, consistent with the existing codebase.

## Risks / Trade-offs

- **[Risk] Scale**: On clusters with thousands of indices, `GET /_all/_settings` can be slow or heavy.
  → **Mitigation**: The request is scoped to a single setting key (`index.lifecycle.name`), which keeps response sizes small. ES handles this efficiently. Future enhancement could add pagination/wildcard targets if profiling shows it is a bottleneck.

- **[Risk] Race condition**: Between the scan and the clear, a new index with the policy could be created.
  → **Mitigation**: The subsequent `DELETE` will still fail with an updated message. This is the same race window that exists today for users doing manual cleanup; we have not made it worse.

- **[Risk] Fleet-managed indices**: Explicitly mutating settings on Fleet-managed data-stream backing indices via `PUT /_settings` could interfere with Fleet's expectations.
  → **Mitigation**: Nulling `index.lifecycle.name` is a safe operation that only removes ILM management. The index data, mappings, and aliases are untouched. Fleet does not rely on ILM policy references to function.

- **[Risk] Closed indices**: `PUT /_settings` on closed indices may return an error.
  → **Mitigation**: `GET /_all/_settings` includes closed indices by default. The `PUT` call is issued to all matched indices. If closed indices fail, the subsequent `DELETE` will report the still-referencing indices. This is acceptable because the root cause is outside Terraform.
