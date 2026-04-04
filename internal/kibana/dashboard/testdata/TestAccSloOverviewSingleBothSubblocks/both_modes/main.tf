variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Invalid: both single and groups configured"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kuery"
  query_text     = ""

  panels = [{
    type = "slo_overview"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    slo_overview_config = {
      single = {
        slo_id = "my-slo-id"
      }
      groups = {
        title = "SLO Groups"
      }
    }
  }]
}
