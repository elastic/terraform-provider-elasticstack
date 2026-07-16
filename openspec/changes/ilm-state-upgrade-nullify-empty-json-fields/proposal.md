# Proposal: ILM state upgrade — nullify empty-string JSON fields

## Problem

Upgrading `elasticstack_elasticsearch_index_lifecycle` from a pre-0.14.4 provider (SDKv2) to any
0.14.4+ provider (Plugin Framework) fails with a fatal `Invalid JSON String Value` error for any ILM
policy whose SDKv2 state stored an empty string for a JSON attribute:

- Top-level `metadata` — omitted in HCL but serialised by SDKv2 as `""`.
- `allocate.include`, `allocate.exclude`, `allocate.require` — each an optional JSON-object
  attribute that SDKv2 serialised as `""` when not set.

```
Error: Invalid JSON String Value
  with elasticstack_elasticsearch_index_lifecycle.example
  A string value was provided that is not valid JSON string format (RFC 7159).
  Given Value: (empty)
```

**Root cause:** The ILM v0→v1 state upgrader (`internal/elasticsearch/index/ilm/state_upgrade.go`)
unwraps singleton-list phase and action blocks but never calls `stateutil.NullifyEmptyString` on
JSON-string attributes. The conversion in `value_conv.go` then passes `""` straight into
`jsontypes.NewNormalizedValue`, which `jsontypes.NormalizedType`'s validator rejects as invalid JSON.

This is the same class of bug fixed for `elasticstack_elasticsearch_index_template` and
`elasticstack_elasticsearch_component_template` in #3914, but that fix did not cover the ILM
resource.

## Recommendation

In `migrateILMStateV0ToV1`, add `stateutil.NullifyEmptyString` calls after the list-unwrapping loop:

1. Call `stateutil.NullifyEmptyString(stateMap, "metadata")` at the top level.
2. For each phase block, after `unwrapPhaseActionLists`, inspect the `allocate` action object (if
   present) and call `stateutil.NullifyEmptyString(allocateObj, "include", "exclude", "require")`.

This mirrors the pattern used by the template and component-template state upgraders and by the
transform resource upgrader (`internal/elasticsearch/transform/state_upgrade.go`).

## Scope

- `internal/elasticsearch/index/ilm/state_upgrade.go` — add `NullifyEmptyString` calls for
  top-level `metadata` and for `allocate.include`, `allocate.exclude`, `allocate.require` within
  each phase.
- `internal/elasticsearch/index/ilm/state_upgrade_test.go` — add unit test cases for the
  `metadata: ""` path and for each of the three allocate JSON attributes as empty strings.
- Acceptance test — add a test covering the SDK → Plugin Framework upgrade path for a policy that
  has never had `metadata` set.
- `openspec/specs/elasticsearch-index-lifecycle/spec.md` — extend REQ-030–REQ-031 to specify that
  the upgrader MUST nullify empty-string JSON attributes.

## Out of scope

- Changes to `stateutil.NullifyEmptyString` itself — the helper already handles empty-string keys
  correctly and is idempotent.
- Changes to `value_conv.go` — the empty-string check belongs in the upgrader, not the converter.
- Any Elasticsearch API or schema changes.
- Changes to `elasticstack_elasticsearch_index_template` or
  `elasticstack_elasticsearch_component_template` — already fixed in #3914.
