# Proposal: ML Anomaly Swim Lane and Single Metric Viewer Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners managing ML anomaly detection jobs through this provider cannot declaratively compose the resulting visualisations on dashboards without falling back to raw `config_json`. The two panel types addressed here — `ml_anomaly_swimlane` and `ml_single_metric_viewer` — are part of Kibana's ML panel family alongside `ml_anomaly_charts` (tracked separately in #4000). All three have first-class typed API models in the generated `kbapi` client. Without typed panel blocks, users must serialize the entire panel configuration as an opaque JSON string, losing schema validation, plan-time feedback, and HCL readability.

## What Changes

Two new typed panel config blocks are added to `elasticstack_kibana_dashboard`, following the panelkit conventions established for other typed panels:

- **`ml_anomaly_swimlane_config`** for panels with `type = "ml_anomaly_swimlane"`.
- **`ml_single_metric_viewer_config`** for panels with `type = "ml_single_metric_viewer"`.

Both handlers are implemented as new packages under `internal/kibana/dashboard/panel/` and registered in `panelHandlers` in `registry.go`.

### `ml_anomaly_swimlane_config` schema sketch

```hcl
panels = [
  {
    type       = "ml_anomaly_swimlane"
    panel_grid = { x = 0, y = 0, w = 24, h = 15 }
    ml_anomaly_swimlane_config = {
      swimlane_type = "viewBy"          # required: "overall" | "viewBy"
      job_ids       = ["high-cpu-detection"]   # required list of strings
      view_by       = "host.name"       # required when swimlane_type = "viewBy"; forbidden otherwise
      per_page      = 10                # optional float
      title         = "Host-level anomalies"
      time_range = {
        from = "now-24h"
        to   = "now"
      }
    }
  }
]
```

- Flat schema with `swimlane_type` as the discriminator (`"overall"` | `"viewBy"`).
- `view_by` is required when `swimlane_type = "viewBy"` and forbidden when `swimlane_type = "overall"`. Enforced by plan-time validators.
- `job_ids` (required `list(string)`, at least one entry).
- `per_page` (optional float — `float32` in the API).
- `title`, `description`, `hide_title`, `hide_border` reuse `panelkit.PanelPresentationAttributes()`.
- `time_range` reuses `panelkit.TimeRangeSchema()`.

### `ml_single_metric_viewer_config` schema sketch

```hcl
panels = [
  {
    type       = "ml_single_metric_viewer"
    panel_grid = { x = 0, y = 15, w = 24, h = 15 }
    ml_single_metric_viewer_config = {
      job_ids                 = ["airline-metric-detection"]  # required list, length-1 validator
      selected_detector_index = 0          # optional float
      forecast_id             = "fc-2026-06-15"   # optional string
      function_description    = "mean"     # optional string
      selected_entities = {
        airline     = { string_value  = "AAL" }
        region_code = { numeric_value = 4 }
      }
      title = "Airline metric — AAL, region 4"
      time_range = {
        from = "now-7d"
        to   = "now"
      }
    }
  }
]
```

- `job_ids` (required `list(string)`, length-1 validator — API shape consistent with sibling ML panels).
- `selected_detector_index` (optional float — `float32` in the API).
- `forecast_id`, `function_description` (optional strings).
- `selected_entities` (`MapNestedAttribute` keyed by field name; each value is an object with mutually exclusive `string_value` and `numeric_value`; plan-time validator enforces exactly one per entry).
- `title`, `description`, `hide_title`, `hide_border`, `time_range` as standard panelkit passthroughs.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-dashboard`: Add two new typed panel config blocks — `ml_anomaly_swimlane_config` (REQ-047) and `ml_single_metric_viewer_config` (REQ-048) — with full create/read/update mapping, plan-time validation, and acceptance-test expectations.

## Impact

- **Specs**: Delta under `openspec/changes/kibana-dashboard-ml-swimlane-smv/specs/kibana-dashboard/spec.md` until merged into canonical spec.
- **Implementation** (future): two new handler packages (`internal/kibana/dashboard/panel/mlanomalyswimlane/`, `internal/kibana/dashboard/panel/mlsinglemetricviewer/`), new model types in `internal/kibana/dashboard/models/`, `registry.go` additions, schema additions in `internal/kibana/dashboard/schema.go`, acceptance tests.
- **Additive only**: no changes to existing panel types or their schemas.
