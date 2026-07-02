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

- `severity` (string enum: `low`, `warning`, `minor`, `major`, `critical`): expands to canonical `{min, max}` at write time. On read, the form is recovered from prior state, not inferred from the API (see "Form preservation" below).
- `min` (int64, required when `severity` is absent) plus optional `max` (int64): raw numeric range; written to API verbatim.

Plan-time validators enforce the "exactly one of `severity` | `min`" constraint. `max` may be set only when `min` is set and `severity` is unset; `severity` together with `max` is a plan-time error. This is enforced via an `objectvalidator.ExactlyOneOf`-style cross-attribute validator on each list item.

**Form preservation (REQ-009 extension):** The API encodes `severity_threshold` as five typed structs (`KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0` through `SeverityThreshold4`, in `generated/kbapi/kibana.gen.go`), each a canonical band carrying only `{min, max}`. It conveys no information about whether the practitioner authored a named `severity` or a raw range. Therefore the read path does **not** infer the form from the API; it preserves the form held in prior state, reusing the existing REQ-009 seeding mechanism (`panelkit.SimpleFromAPI` threads `prior` into `PopulateFromAPI`; the new `severityThresholdFromAPI(apiItem, priorItem)` helper mirrors `sloStringFromAPIOrPrior`):

- **prior named** (`prior.Severity` known, `Min`/`Max` null): store named form; resolve the label from the API `{min,max}` via the canonical-band table. If the API value no longer matches any canonical band, fall back to raw `min`/`max` (surfacing as drift).
- **prior raw** (`prior.Min`/`Max` known, `Severity` null): store raw `min`/`max` verbatim — do **not** coerce to a named severity even when the pair equals a canonical band.
- **import** (`prior == nil`): no prior form exists to preserve; default to named form when the API `{min,max}` matches a canonical band, else raw.

This guarantees a stable configuration produces no plan diff. Normalization to named form is permitted **only on import**.

Anomaly scores are integers in `[0, 100]`; the five canonical bands match the generated Kibana OpenAPI const values (`KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0`–`SeverityThreshold4`):

| `severity` value | API `min` | API `max`   |
|---|---|---|
| `low`      | 0  | 3            |
| `warning`  | 3  | 25           |
| `minor`    | 25 | 50           |
| `major`    | 50 | 75           |
| `critical` | 75 | — (open-ended, no `max`) |

On **write**: named severity → expand to canonical pair; raw range → use as-is.
On **read**: the form (named vs raw) is recovered from prior state, not inferred from the API `{min, max}` pair (see "Form preservation" above). When the prior item is named, the label is derived from the API pair via the canonical-band table; when the prior item is raw, the API pair is stored verbatim. The `critical` band has no `max` in the API payload; a `{min: 75}` item whose prior form is named maps to `severity = "critical"`, while a `{min: 75}` item whose prior form is raw maps to `min = 75` with `max` null.

**`job_ids` remains a plain string list (per human direction #2b):**
No plan-time existence check. Matches the provider-wide pattern for cross-resource references.

**`time_range` reuses `panelkit.TimeRangeSchema`:**
The same helper is used by `discoversession` and other panel types. Only `from` and `to` are sent/read; `mode` follows REQ-009 null-preservation.

**API dispatch:**
The model layer serializes `KibanaHTTPAPIsMlAnomalyCharts` directly. The `SeverityThreshold` field is `*[]KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item` (a union type with `json.RawMessage` internals). The implementation must marshal each item using the concrete structs (`SeverityThreshold0`–`SeverityThreshold4` for named bands, or a raw `{min, max?}` struct for custom ranges).

**Package location:** `internal/kibana/dashboard/panel/mlanomalycharts/` — consistent with the `sloburnrate`, `syntheticsmonitors`, and `discoversession` siblings.

## Risks / Trade-offs

- [Risk] The `SeverityThreshold_Item` union uses `json.RawMessage` internally; marshaling requires constructing the correct concrete struct and marshaling to JSON. The five canonical structs have fixed-point enum values, so construction is deterministic. Custom ranges use a plain `struct{ Min int; Max *int }`. → Mitigation: implement a `buildSeverityThresholdItem` helper that takes `severity *string, min *int64, max *int64` and returns the correct raw JSON payload.
- [Trade-off] The API carries no information about whether a `severity_threshold` item was authored as a named `severity` or a raw `{min, max}` range. The provider resolves this form from prior state on read (REQ-009 extension) so that a stable configuration produces no plan diff; normalization to named form is permitted only on import, where no prior form exists. This is consistent with how the provider preserves representation facets (e.g. `time_range.mode`) elsewhere.
- [Assumption] The `KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema` shape (from/to, no mode in the API) is handled correctly by reusing `panelkit.TimeRangeSchema` with REQ-009 null-preservation on `mode`. The API does not echo `mode` back; the provider preserves the prior value of `mode` rather than zeroing it.

## Open questions

None. The human direction comment resolved both open design questions (severity threshold shape, `job_ids` validation).
