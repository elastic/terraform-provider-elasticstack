# Typed by-value `metric_chart_config` (no `by_value.config_json`); used for no-drift
# second-apply and import/read checks on panel `config_json` (task 4.3/4.4).
variable "dashboard_title" { type = string }

resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "lens-dashboard-app: typed by_value metric (acc task 4)"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "lens-dashboard-app"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    lens_dashboard_app_config = {
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
