variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with unsupported synthetics_stats_overview config_json"

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
    type = "synthetics_stats_overview"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    config_json = jsonencode({
      title = "My Panel"
    })
  }]
}
