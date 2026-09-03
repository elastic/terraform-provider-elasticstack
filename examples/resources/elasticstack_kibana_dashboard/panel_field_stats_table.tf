// Example: field_stats_table panels — data view (`by_dataview`) and ES|QL (`by_esql`) branches.

resource "elasticstack_kibana_data_view" "field_stats" {
  override = true
  data_view = {
    title          = "field-stats-example-*"
    name           = "field-stats-example"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "with_field_stats_table_panels" {
  title            = "Dashboard with field_stats_table panels"
  description      = "Typed field statistics table panels: by_dataview + by_esql"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [
    {
      type = "field_stats_table"
      grid = { x = 0, y = 0, w = 24, h = 15 }
      field_stats_table_config = {
        by_dataview = {
          data_view_id       = elasticstack_kibana_data_view.field_stats.data_view.id
          show_distributions = true
          title              = "Field statistics — data view"
          time_range = {
            from = "now-24h"
            to   = "now"
          }
        }
      }
    },
    {
      type = "field_stats_table"
      grid = { x = 0, y = 15, w = 24, h = 15 }
      field_stats_table_config = {
        by_esql = {
          query              = "FROM logs-* | LIMIT 100"
          show_distributions = false
          title              = "Field statistics — ES|QL"
        }
      }
    },
  ]
}
