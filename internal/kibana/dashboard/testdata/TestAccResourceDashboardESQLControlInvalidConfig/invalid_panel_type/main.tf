variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with invalid esql_control_config panel type"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kql"
  query_text     = ""

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    esql_control_config = {
      selected_options = []
      variable_name    = "my_var"
      variable_type    = "values"
      esql_query       = "FROM logs-* | STATS count = COUNT(*) BY host.name"
      control_type     = "STATIC_VALUES"
    }
  }]
}
