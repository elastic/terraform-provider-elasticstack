variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Burn Rate Panel (display options)"

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
    slo_burn_rate_config = {
      slo_id      = "test-slo-id"
      duration    = "5m"
      title       = "My Burn Rate Panel"
      description = "Monitors the 5-minute burn rate"
      hide_title  = true
      hide_border = false
    }
  }]
}
