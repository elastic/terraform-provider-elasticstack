variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  panels = [{
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
    }
    config_json = jsonencode({ key = "value" })
  }]
}
