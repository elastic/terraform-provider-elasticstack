variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with unsupported slo_burn_rate config_json"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kql"
  query_text     = ""

  panels = [{
    type = "slo_burn_rate"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    config_json = jsonencode({
      slo_id   = "test-slo-id"
      duration = "72h"
    })
  }]
}
