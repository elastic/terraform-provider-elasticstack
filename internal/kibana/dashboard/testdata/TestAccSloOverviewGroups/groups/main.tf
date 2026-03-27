variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Groups Overview Panel"

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
      groups = {
        title       = "SLO Groups"
        description = "Overview of all SLOs grouped by status"
        group_filters = {
          group_by  = "status"
          kql_query = "slo.name: my-*"
        }
      }
    }
  }]
}
