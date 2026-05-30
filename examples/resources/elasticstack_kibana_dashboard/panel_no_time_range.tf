// Example: typed Lens panel without chart-level time_range — the panel uses the dashboard global time picker.

resource "elasticstack_kibana_dashboard" "metric_no_panel_time_range" {
  title            = "Dashboard with metric panel (no chart time_range)"
  description      = "Omit time_range on the chart block to defer to the dashboard-level window"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 12 }
    vis_config = {
      by_value = {
        metric_chart_config = {
          data_source_json = jsonencode({
            type          = "data_view_spec"
            index_pattern = "metrics-*"
            time_field    = "@timestamp"
          })
          query = { expression = "" }
          metrics = [{
            config_json = jsonencode({
              type      = "primary"
              operation = "count"
              format    = { type = "number" }
            })
          }]
        }
      }
    }
  }]
}
