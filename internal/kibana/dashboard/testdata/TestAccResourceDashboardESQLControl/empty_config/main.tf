variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ES|QL Control Panel (empty config block)"

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
    type = "esql_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
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
