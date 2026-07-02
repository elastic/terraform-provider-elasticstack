## MODIFIED Requirements

### Requirement: config_json unsupported for typed panel types (REQ-010) — partial update

This requirement extends REQ-010 to include the `field_stats_table` panel type. Practitioner-authored panel-level `config_json` SHALL NOT be supported for `field_stats_table` panels; using `config_json` with `type = "field_stats_table"` SHALL return an error diagnostic stating that `config_json` is not supported for `field_stats_table`.

#### Scenario: config_json rejected for field_stats_table panel type

- GIVEN a panel with `type = "field_stats_table"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `field_stats_table`

---

## ADDED Requirements

### Requirement: Field statistics table panel behavior (REQ-047)

When a panel entry sets `type = "field_stats_table"`, the resource SHALL accept a `field_stats_table_config` block and SHALL require that block to be present when the panel type is `field_stats_table`. The block SHALL expose two mutually exclusive sub-blocks mirroring the API's `view_type` discriminated union:

- `by_dataview` — backed by a Kibana data view (API `view_type = "dataview"`).
- `by_esql` — backed by an ES|QL query (API `view_type = "esql"`).

Exactly one of `by_dataview` or `by_esql` SHALL be set; setting both or neither SHALL produce an error diagnostic at plan time.

The `field_stats_table_config` block SHALL conflict with all other typed panel config blocks and with panel-level `config_json`. When `type = "field_stats_table"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `field_stats_table`.

#### The `by_dataview` sub-block

The `by_dataview` sub-block SHALL accept:

- `data_view_id` (required, string) — the identifier of the source data view.
- `show_distributions` (optional, bool) — whether to show distribution mini-charts in the table; null-preserved on read per REQ-009: when prior state has it null, the provider keeps it null even if Kibana returns a server-side default.
- `title` (optional, string) — panel display title; null-preserved on read.
- `description` (optional, string) — panel description; null-preserved on read.
- `hide_title` (optional, bool) — whether to hide the panel title; null-preserved on read.
- `time_range` (optional, object `{ from = required string, to = required string, mode = optional string }`) — panel-level time range override; null-preserved on read: when prior state has `time_range` null, the provider keeps it null even if Kibana returns values.

On write, the provider SHALL set `view_type = "dataview"` internally and map `data_view_id` and optional fields to the API payload. The `view_type` field is not exposed as a user-facing attribute.

#### The `by_esql` sub-block

The `by_esql` sub-block SHALL accept:

- `query` (required, string) — the ES|QL query string; mapped to `query.esql` in the API payload.
- `show_distributions` (optional, bool) — null-preserved on read per REQ-009.
- `title` (optional, string) — null-preserved on read.
- `description` (optional, string) — null-preserved on read.
- `hide_title` (optional, bool) — null-preserved on read.
- `time_range` (optional, object `{ from = required string, to = required string, mode = optional string }`) — null-preserved on read.

On write, the provider SHALL set `view_type = "esql"` internally and map `query` to `query.esql` and optional fields to the API payload.

#### Read behavior

On read, the resource SHALL detect the `view_type` field in the API response and populate the matching sub-block (`by_dataview` or `by_esql`), leaving the other sub-block null. For each optional attribute, the resource SHALL apply REQ-009 null-preservation: if prior state had the attribute null, the provider SHALL keep it null even if the API response contains a value for it.

#### HCL example — by_dataview branch

```hcl
panels = [
  {
    type = "field_stats_table"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    field_stats_table_config = {
      by_dataview = {
        data_view_id       = "logs-view"
        show_distributions = true
        title              = "Field statistics — logs view"
        hide_title         = false
        hide_border        = false
        time_range = {
          from = "now-24h"
          to   = "now"
        }
      }
    }
  }
]
```

#### HCL example — by_esql branch

```hcl
panels = [
  {
    type = "field_stats_table"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    field_stats_table_config = {
      by_esql = {
        query              = "FROM logs | STATS count = COUNT(*) BY service.name"
        show_distributions = true
        title              = "Field statistics — logs by service"
        time_range = {
          from = "now-24h"
          to   = "now"
        }
      }
    }
  }
]
```

#### Scenario: by_dataview branch create/read round-trip

- GIVEN a panel with `type = "field_stats_table"` and `field_stats_table_config.by_dataview = { data_view_id = "logs-view", show_distributions = true, time_range = { from = "now-24h", to = "now" } }`
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL contain `by_dataview` populated with the same values, `by_esql` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: by_esql branch create/read round-trip

- GIVEN a panel with `type = "field_stats_table"` and `field_stats_table_config.by_esql = { query = "FROM logs | STATS count = COUNT(*) BY service.name", show_distributions = false }`
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL contain `by_esql` populated with the same values, `by_dataview` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: Exactly one of by_dataview or by_esql

- GIVEN a panel with both `field_stats_table_config.by_dataview` and `field_stats_table_config.by_esql` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `by_dataview` or `by_esql` must be set

#### Scenario: Neither branch set

- GIVEN a panel with `field_stats_table_config = {}` (neither `by_dataview` nor `by_esql` set)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `by_dataview` or `by_esql` must be set

#### Scenario: time_range null-preservation

- GIVEN a `field_stats_table_config.by_dataview` panel whose prior state has `time_range = null`
- WHEN the post-apply read returns a panel where Kibana populated `time_range` with values
- THEN state SHALL keep `time_range` null and a subsequent plan against configuration that omits `time_range` SHALL show no changes

#### Scenario: show_distributions null-preservation

- GIVEN a `field_stats_table_config.by_esql` panel whose prior state has `show_distributions = null`
- WHEN the post-apply read returns a panel where Kibana populated `show_distributions`
- THEN state SHALL keep `show_distributions` null and a subsequent plan against configuration that omits it SHALL show no changes

This behavior is already covered by the REQ-010 update above.

#### Scenario: Drift detection — Kibana returns branch data intact

- GIVEN an existing dashboard with a `field_stats_table` panel in state
- WHEN Kibana returns the same panel configuration on a subsequent read
- THEN a plan SHALL show no changes
