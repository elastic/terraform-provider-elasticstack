# Plan check: both config_json and a typed by-value chart are set; validator must reject (task 3.1).
variable "dashboard_title" { type = string }
resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "plan: by_value config_json + metric_chart_config"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
  panels = [{
    type = "lens-dashboard-app"
    grid = { x = 0, y = 0, w = 4, h = 4 }
    lens_dashboard_app_config = {
      by_value = {
        config_json = jsonencode({ type = "metric" })
        metric_chart_config = {
          data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
          query            = { expression = "" }
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
