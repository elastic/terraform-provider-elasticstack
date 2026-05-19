# Plan check: by_value is set (so outer lens_dashboard_app_config is valid) but
# `lensDashboardAppByValueSourceValidator` rejects an empty by_value (REQ task 3.2).
variable "dashboard_title" { type = string }
resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "plan: by_value with no config_json or typed chart"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
  panels = [{
    type = "lens-dashboard-app"
    grid = { x = 0, y = 0, w = 4, h = 4 }
    lens_dashboard_app_config = {
      by_value = {}
    }
  }]
}
