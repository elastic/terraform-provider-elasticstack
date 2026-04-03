variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Error Budget Panel (no instance ID)"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kuery"
  query_text     = ""

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
      # slo_instance_id is intentionally omitted to verify null-preservation
    }
  }]
}
