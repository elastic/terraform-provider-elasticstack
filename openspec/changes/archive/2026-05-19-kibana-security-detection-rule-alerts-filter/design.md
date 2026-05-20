## Context

`elasticstack_kibana_security_detection_rule` actions expose an `alerts_filter` attribute typed as `schema.MapAttribute{ElementType: types.StringType}` (see `internal/kibana/security_detection_rule/schema.go` lines 403–407 and `models.go` line 184: `AlertsFilter types.Map`).

The Kibana Detection Engine API expects `alertsFilter` as a nested JSON object with a `query` sub-object (containing `kql` string and `filters` array) and an optional `timeframe` object. The current flat map cannot represent this shape.

- **Read bug** (`models_from_api_type_utils.go` lines 194–205): iterates `*apiAction.AlertsFilter` (a `map[string]interface{}`) and calls `fmt.Sprintf("%v", v)` for each value, producing Go map literal strings instead of JSON.
- **Write bug** (`models_to_api_type_utils.go` lines 559–570): re-assembles the flat string map into `map[string]any` and sends as the API payload; nested values arrive as plain strings that the API rejects with a 400 error.

The `elasticstack_kibana_alerting_rule` resource already implements `alerts_filter` as a structured `SingleNestedBlock` with `kql` and a `timeframe` sub-block (added in the flapping change; see `internal/kibana/alertingrule/schema.go` lines 260–311 and `models.go` lines 559–678 for the reference implementation).

The Detection Engine API client type `SecurityDetectionsAPIRuleActionAlertsFilter` is `map[string]interface{}` (see `generated/kbapi/kibana.gen.go` lines 54824–54836), while the alerting rule uses typed `models.ActionAlertsFilter` structs. Serialization for the detection rule must convert the structured Terraform model to/from the untyped map.

## Goals

- Fix the read and write paths for `alerts_filter` in `elasticstack_kibana_security_detection_rule` by replacing the broken `MapAttribute` with a `SingleNestedBlock`.
- Model `query.kql` as a typed string attribute and `query.filters_json` as a `jsontypes.Normalized` JSON string (named `filters_json` to allow a future typed `filters` attribute to be added).
- Include `timeframe` sub-block with `days`, `timezone`, `hours_start`, `hours_end` — matching the `elasticstack_kibana_alerting_rule` pattern for consistency.
- Bump schema version 1 → 2 with a no-op `StateUpgrade` function (the feature was always broken; no valid state to migrate).

## Non-Goals

- Full structured modelling of individual Kibana filter DSL objects inside `query.filters` (follow-up).
- Changes to `elasticstack_kibana_alerting_rule` (already correct).
- Changes to `response_actions`.

## Decisions

| Topic | Decision |
|-------|----------|
| Terraform shape | `SingleNestedBlock` at `actions.alerts_filter` with nested `query` block and optional `timeframe` block |
| `query.filters` name | `filters_json` — reserves `filters` for a future typed list; consistent with human direction on the issue |
| `query.filters_json` type | `jsontypes.Normalized` — same pattern as `params` in the same resource |
| `query.kql` | Optional string; absent when not configured |
| `timeframe` | Optional nested block identical to `elasticstack_kibana_alerting_rule`: `days` (list int64), `timezone` (string), `hours_start` (string), `hours_end` (string) |
| Timeframe validation | All four attributes required when the block is present (via `objectvalidator.AlsoRequires` or equivalent) |
| API serialization (write) | Unmarshal `filters_json` JSON string into `[]interface{}`, build `map[string]interface{}{"query": {"kql": ..., "filters": [...]}, "timeframe": {...}}`, cast to `kbapi.SecurityDetectionsAPIRuleActionAlertsFilter` |
| API deserialization (read) | Extract nested `query` and `timeframe` maps from `*apiAction.AlertsFilter`, marshal sub-values to normalized JSON for `filters_json`, extract typed fields for `kql` and timeframe |
| Schema version | Bump 1 → 2; `StateUpgraders` entry with a no-op function that discards the old broken `MapAttribute` state |
| New model structs | `ActionAlertsFilterModel`, `ActionAlertsFilterQueryModel` (parallel to alertingrule's `alertsFilterModel`/`timeframeModel`) |

## Risks / Trade-offs

- **Breaking change**: any practitioner who has `alerts_filter` in their configuration must migrate to the nested block syntax. The feature was non-functional before, so breakage is in configuration syntax only — no data loss risk.
- **Untyped API client**: `SecurityDetectionsAPIRuleActionAlertsFilter` is `map[string]interface{}`, requiring manual marshaling. This is isolated to the expand/flatten helpers.

## Open questions

- _(Resolved by human direction)_ `filters` attribute name: use `filters_json`.
- _(Resolved by human direction)_ `timeframe` in scope: yes.
- _(Resolved by human direction)_ State migration: no-op (broken state only).
