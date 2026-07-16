## Context

`elasticstack_kibana_alerting_rule` migrated from the SDKv2 provider to the Plugin
Framework provider in a previous release. The state upgrader (`migrateV0ToV1` in
`internal/kibana/alertingrule/state_upgrade.go`) already normalises several
SDKv2-vs-Plugin-Framework differences:

- Plain string fields `notify_when` and `throttle` are nullified when empty
  (`stateutil.NullifyEmptyString`).
- `actions[].frequency`, `actions[].alerts_filter`, and
  `actions[].alerts_filter.timeframe` are collapsed from SDKv2 list shape to Plugin
  Framework object shape (`stateutil.CollapseListPath`).

However, neither the rule-level `params` nor the per-action `params` fields are
normalised. Both are typed as `jsontypes.Normalized` in the Plugin Framework model.
The Plugin Framework validates any `jsontypes.Normalized` value against JSON syntax
at decode time; an empty string fails that validation.

The SDKv2 provider stored empty-string values for optional string attributes when
not configured. A `params` left unconfigured (or explicitly configured as `""`)
in the old provider would therefore be stored as `""` in state. After upgrading to
the Plugin Framework provider and running `terraform plan`, decoding that state
raises `Invalid JSON String Value`.

The same pattern was fixed for:
- `elasticstack_elasticsearch_index_lifecycle` (#3914 / #3855 class)
- `elasticstack_elasticsearch_index_template` and related resources

The fix follows the identical pattern: call `stateutil.NullifyEmptyString` for the
affected JSON-typed keys during state upgrade.

## Goals

- Prevent `Invalid JSON String Value` errors for practitioners upgrading from
  SDKv2 provider builds with empty-string `params` in state.
- Bring `elasticstack_kibana_alerting_rule`'s state upgrader into parity with the
  pattern established for ILM and template resources.

## Non-Goals

- Changing the alerting rule API shape, schema, or any user-facing behaviour.
- Fixing any other structural issue that may affect state upgrade (e.g. the
  "missing expected {" list-vs-object shape warning noted in #4147 — this is a
  separate concern and out of scope).

## Decisions

| Topic | Decision |
|-------|-----------|
| Rule-level `params` | Add `stateutil.NullifyEmptyString(stateMap, "params")` after the existing `stateutil.NullifyEmptyString(stateMap, "notify_when", "throttle")` call. |
| Per-action `params` | Inside the existing `actions` loop, add `stateutil.NullifyEmptyString(action, "params")` alongside the existing `CollapseListPath` calls. |
| Call site ordering | Nullify `params` before the `CollapseListPath` calls within the action loop to be consistent with the rule-level ordering (normalise scalars first). |
| Unit tests | Add a dedicated state-upgrade unit test covering: (a) rule-level `params` empty string, (b) action-level `params` empty string, (c) both simultaneously, (d) non-empty / null `params` unchanged. |
| No state version bump | The fix targets the existing v0 → v1 migration only; no new version is introduced. |

## Risks / Trade-offs

- **Existing v1 state**: Practitioners who are already on the Plugin Framework
  provider and have valid JSON in `params` are unaffected — `NullifyEmptyString`
  only acts when the value is exactly `""`.
- **Over-nullification**: Setting `params` to null when it was legitimately `""`
  is safe: the Plugin Framework model treats `null params` as no params configured,
  which is the correct semantic when the SDKv2 stored an empty string to represent
  "not configured".

## Open Questions

_(none — the fix is well-understood by analogy with existing ILM and template
state upgrader fixes)_

## Migration / State

No state version increment is needed. The existing v0 → v1 upgrader is extended in
place; its semantics remain compatible with all v0 states that do not have
empty-string `params`.
