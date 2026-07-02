## Context

The Kibana Dashboard API defines `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts` (in `generated/kbapi/kibana.gen.go` near line 47273) and its config struct `KibanaHTTPAPIsMlAnomalyCharts` (near line 51092). The config is a **flat struct** (not a union), which maps cleanly to the panelkit `PanelPresentationAttributes()` + custom attribute pattern used by `sloburnrate`, `syntheticsmonitors`, and `discoversession`.

The only non-trivial design point is `severity_threshold`: the API defines it as a union of five named-range structs (`KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0` through `SeverityThreshold4`), each corresponding to a Kibana ML canonical severity band. The Terraform schema exposes this as a list of `severity_threshold` items where each item uses either a named enum shortcut or a raw numeric range escape hatch, but not both.

## Goals / Non-Goals

**Goals:**
- Full typed coverage of the `ml_anomaly_charts` panel config, with plan-time validation.
- Named severity shortcut (`severity` string enum) as the primary UX path; raw `{min, max}` as the escape hatch.
- Null-preservation (REQ-009) for all optional fields.
- Reject `config_json` on `ml_anomaly_charts` panels (REQ-010).
- Follow the existing `sloburnrate`/`syntheticsmonitors` flat-config panelkit pattern exactly.

**Non-Goals:**
- Cross-referencing `job_ids` against Kibana's ML API at plan time. Job IDs are opaque strings; invalid IDs surface as Kibana API errors on apply (consistent with `slo_id` treatment on SLO panels and `data_view_id` on control panels).
- Related ML panel types (`ml_anomaly_swimlane`, `ml_single_metric_viewer`) — tracked separately.

## Decisions

**Severity threshold union shape (per human direction #2a):**
Each `severity_threshold` list item must set exactly one of:
- `severity` (string enum: `low`, `warning`, `minor`, `major`, `critical`): expands to canonical `{min, max}` at write time, collapses back to enum name on read.
- `min` (int64, required when `severity` is absent) plus optional `max` (int64): raw numeric range; written to API verbatim.

Plan-time validators enforce the "exactly one of `severity` | `min`" constraint. `max` MAY be set only when `min` is set (and `severity` is unset). This is enforced via an `objectvalidator.ExactlyOneOf`-style cross-attribute validator on each list item.

Anomaly scores are integers in `[0, 100]`; the five canonical bands are:

| `severity` value | API `min` | API `max`   |
|---|---|---|
| `low`      | 0  | 2  (inclusive) |
| `warning`  | 3  | 24 (inclusive) |
| `minor`    | 25 | 49 (inclusive) |
| `major`    | 50 | 74 (inclusive) |
| `critical` | 75 | — (open-ended, no `max`) |

On **write**: named severity → expand to canonical pair; raw range → use as-is.
On **read**: `{min, max}` pair → check against canonical bands; if it matches, store as named severity; otherwise store as raw range with `min`/`max`.
The `critical` band has no `max` in the API payload; on read a `{min: 75}` item maps to `severity = "critical"`.

This round-trips correctly for the common case (named severity → apply → read → plan shows no diff). A raw range that happens to match a canonical band round-trips to the named form (acceptable — the canonical form is the preferred representation).

**`job_ids` remains a plain string list (per human direction #2b):**
No plan-time existence check. Matches the provider-wide pattern for cross-resource references.

**`time_range` reuses `panelkit.TimeRangeSchema`:**
The same helper is used by `discoversession` and other panel types. Only `from` and `to` are sent/read; `mode` follows REQ-009 null-preservation.

**API dispatch:**
The model layer serializes `KibanaHTTPAPIsMlAnomalyCharts` directly. The `SeverityThreshold` field is `*[]KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item` (a union type with `json.RawMessage` internals). The implementation must marshal each item using the concrete structs (`SeverityThreshold0`–`SeverityThreshold4` for named bands, or a raw `{min, max?}` struct for custom ranges).

**Package location:** `internal/kibana/dashboard/panel/mlanomalycharts/` — consistent with the `sloburnrate`, `syntheticsmonitors`, and `discoversession` siblings.

## Risks / Trade-offs

- [Risk] The `SeverityThreshold_Item` union uses `json.RawMessage` internally; marshaling requires constructing the correct concrete struct and marshaling to JSON. The five canonical structs have fixed-point enum values, so construction is deterministic. Custom ranges use a plain `struct{ Min int; Max *int }`. → Mitigation: implement a `buildSeverityThresholdItem` helper that takes `severity *string, min *int64, max *int64` and returns the correct raw JSON payload.
- [Risk] On read, a custom range `{min: 3, max: 24}` coincidentally matches the `warning` band and will be returned as `severity = "warning"`. This is acceptable behavior (canonical form preferred) and should be documented in the spec.
- [Assumption] The `KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema` shape (from/to, no mode in the API) is handled correctly by reusing `panelkit.TimeRangeSchema` with REQ-009 null-preservation on `mode`. The API does not echo `mode` back; the provider preserves the prior value of `mode` rather than zeroing it.

## Open questions

None. The human direction comment resolved both open design questions (severity threshold shape, `job_ids` validation).
