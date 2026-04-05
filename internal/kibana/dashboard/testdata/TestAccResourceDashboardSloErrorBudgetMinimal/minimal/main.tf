variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Error Budget Panel (minimal)"

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
    type = "slo_error_budget"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 12
    }
    slo_error_budget_config = {
      slo_id = "my-slo-id"
    }
  }]
}
