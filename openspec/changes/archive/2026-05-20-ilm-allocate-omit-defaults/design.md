## Context

Canonical requirements for the ILM resource live in [`openspec/specs/elasticsearch-index-lifecycle/spec.md`](../../specs/elasticsearch-index-lifecycle/spec.md). Implementation lives in [`internal/elasticsearch/index/ilm/`](../../../internal/elasticsearch/index/ilm/).

The Elasticsearch ILM Allocate action API treats all five parameters (`number_of_replicas`, `total_shards_per_node`, `include`, `exclude`, `require`) as independently optional; at least one must be present. Omitting `number_of_replicas` from a PUT request leaves the current replica setting unchanged on the index.

## Goals / Non-Goals

**Goals:**

- Remove provider-injected default values for `number_of_replicas` and `total_shards_per_node` in the `allocate` block so omitting them in configuration means they are not sent to the Elasticsearch API.
- Fix the flatten-side bug that forces `total_shards_per_node = -1` from the API response even when the field is absent.
- Preserve full round-trip fidelity: when a user (or an imported policy) has explicit values for these fields, they are correctly read back from the API into state and appear in the plan.
- Ensure no plan diff on upgrade for existing resources that have `0` / `-1` in state (use `UseStateForUnknown`).
- Update the `number_of_replicas` description string to remove the inaccurate "Default: `0`" text.

**Non-goals:**

- Emitting a deprecation/upgrade warning for users whose state has the old injected default values (explicitly decided: No).
- Adding a CHANGELOG or upgrade guide entry (explicitly decided: No).
- Fixing similar Default-injection patterns in other ILM actions (none identified as exhibiting this pattern).
- Retroactively nulling out `0` / `-1` in existing state via a state upgrade (distinguishing user-explicit `0` from provider-injected `0` is not possible from state alone).
- Changing `include`, `exclude`, or `require` field behavior (already correctly `Optional: true` without `Computed` or `Default`).

## Decisions

- **`UseStateForUnknown`** plan modifier on both fields (rather than removing `Computed: true`) to preserve import and read-back fidelity. Without `Computed: true`, fields absent from an API response cannot be stored in state, breaking `terraform import` for policies that have explicit replica counts.
- **No state schema migration**: `UseStateForUnknown` bridges old state (explicit `0` / `-1`) to new schema (no `Default`) without a breaking plan diff. A state migration is not warranted because there is no safe way to distinguish user-explicit `0` from provider-injected `0`.
- **`expand.go` cleanup**: The `def: -1` on `total_shards_per_node` in `ilmActionSettingOptions` is dead code (the version-gate path only executes when `minVersion != nil`; `total_shards_per_node` has no `minVersion`). Remove it to avoid misleading future readers. `skipEmptyCheck: true` is retained so that explicit values of `0` and `-1` are still forwarded to the API.

## Risks / Trade-offs

- **Existing resources with injected defaults**: After upgrade, `terraform plan` will show no diff (due to `UseStateForUnknown`). However, the API payload on the next apply will still contain `number_of_replicas=0` / `total_shards_per_node=-1` because those values remain in state. Self-healing only occurs if the user destroys and recreates the resource, or manually removes those values from the API and re-imports. This is an acceptable trade-off: behavior does not regress.
- **`skipEmptyCheck: true`**: Without this flag, an explicit `total_shards_per_node = 0` would be filtered out by `typeutils.IsEmpty`. The flag is intentional and must be preserved.

## Open Questions

- None. The following questions were answered by the issue author:
  - Q1: No deprecation/upgrade warning needed.
  - Q2: Yes — update `number_of_replicas` description to remove "Default: `0`" language.
  - Q3: No CHANGELOG entry needed.
