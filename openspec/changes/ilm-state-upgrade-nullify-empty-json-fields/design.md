# Design: ILM state upgrade — nullify empty-string JSON fields

## Overview

The ILM state upgrader `migrateILMStateV0ToV1`
(`internal/elasticsearch/index/ilm/state_upgrade.go`) unwraps singleton-list phase and action
blocks from SDKv2-shaped state but never normalises empty-string JSON attributes to `null`. The
Plugin Framework's `jsontypes.NormalizedType` validator rejects `""` as invalid JSON (RFC 7159),
causing a fatal error during the first `terraform plan` after a provider upgrade.

The fix is additive: call `stateutil.NullifyEmptyString` for each JSON-string attribute that SDKv2
could have serialised as `""`. No other logic changes are required.

## JSON-string attributes affected

| Attribute path | Notes |
|---|---|
| `metadata` (top-level) | Optional JSON object; SDKv2 writes `""` when omitted |
| `<phase>.allocate.include` | Optional JSON object; SDKv2 writes `""` when omitted |
| `<phase>.allocate.exclude` | Optional JSON object; SDKv2 writes `""` when omitted |
| `<phase>.allocate.require` | Optional JSON object; SDKv2 writes `""` when omitted |

Phases that support `allocate`: `warm`, `cold`.

## Change details

### `internal/elasticsearch/index/ilm/state_upgrade.go`

Add two targeted calls inside `migrateILMStateV0ToV1` during the v0 → v1 upgrade:

~~~go
for _, pk := range ilmPhaseBlockKeys {
    resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, pk, pk)...)
    if resp.Diagnostics.HasError() {
        return
    }
    if phaseObj, ok := stateMap[pk].(map[string]any); ok {
        unwrapPhaseActionLists(phaseObj)
        if allocateObj, ok := phaseObj[ilmActionAllocate].(map[string]any); ok {
            stateutil.NullifyEmptyString(allocateObj, attrInclude, attrExclude, attrRequire)
        }
    }
}
stateutil.NullifyEmptyString(stateMap, "metadata")
~~~

`stateutil.NullifyEmptyString` is idempotent: if a key is absent, already `nil`, or already a
non-empty string, it is a no-op.

## Testing strategy

### Unit tests (`internal/elasticsearch/index/ilm/state_upgrade_test.go`)

Add the following test cases to the existing `migrateILMStateV0ToV1` unit test table:

1. **`metadata_empty_string`** — top-level `"metadata": ""`, no phases. Assert `metadata == nil`
   after upgrade, and that the upgraded state decodes against the v1 schema without error.

2. **`allocate_include_empty_string`** — a warm phase with an `allocate` block where `"include": ""`
   (and no `exclude`/`require`). Assert `warm.allocate.include == nil` after upgrade.

3. **`allocate_all_json_attrs_empty_string`** — a warm phase `allocate` block where `"include": ""`,
   `"exclude": ""`, and `"require": ""` are all set. Assert all three are `nil` after upgrade.

4. **`metadata_and_allocate_empty_strings`** — combination of `"metadata": ""` at top level and a
   cold phase `allocate` with all three JSON attributes empty. Assert all four fields become `nil`.

### Acceptance tests

Add `TestAccResourceILMFromSDK` (or a separate `TestAccResourceILMFromSDKNoMetadata` variant) to
cover:

- SDK 0.14.5 provider creation of an ILM policy with no `metadata` and a warm phase with an
  `allocate` block that omits `include`/`exclude`/`require`.
- Provider upgrade to the current Plugin Framework version.
- `terraform plan` produces no diff (no-op plan).

This test requires a live Elastic Stack (acceptance test gate, not unit test).

### OpenSpec spec update

Extend `openspec/specs/elasticsearch-index-lifecycle/spec.md` at REQ-030–REQ-031 with an additional
scenario specifying that the upgrader normalises empty-string JSON fields.

## Open questions

_None — the scope is well-defined by the issue and the existing pattern from #3914._

## Invariants

- `stateutil.NullifyEmptyString` is idempotent and safe to call on any map.
- The fix applies once, during the first plan after the provider upgrade. Subsequent plans operate on
  Plugin Framework-written state, which always emits `null` for unset JSON strings.
- No changes to `stateutil.NullifyEmptyString` itself are needed.
- The `ilmActionAllocate` and `attrInclude`/`attrExclude`/`attrRequire` constants already exist in
  `internal/elasticsearch/index/ilm/constants.go`.
