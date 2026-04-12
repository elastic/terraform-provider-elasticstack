variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with synthetics_stats_overview_config on wrong panel type"

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
    type = "slo_burn_rate"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    synthetics_stats_overview_config = {
      title = "Wrong panel type"
    }
  }]
}
