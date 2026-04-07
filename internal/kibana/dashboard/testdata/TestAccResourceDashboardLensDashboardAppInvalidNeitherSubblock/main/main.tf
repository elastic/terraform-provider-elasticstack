variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
    }
    lens_dashboard_app_config = {
      title = "No sub-block set"
    }
  }]
}
