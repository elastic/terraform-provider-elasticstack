variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with missing slo_burn_rate_config"

  time_range {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval {
    pause = true
    value = 0
  }
  query {
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
  }]
}
