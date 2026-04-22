variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Burn Rate Panel (with slo_instance_id)"

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
      slo_id          = "test-slo-id"
      duration        = "6d"
      slo_instance_id = "host-a"
      title           = "Burn Rate: host-a"
      description     = "Monitors the 6-day burn rate for host-a"
    }
  }]
}
