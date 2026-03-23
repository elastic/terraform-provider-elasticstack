variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with ES|QL VALUES_FROM_QUERY control"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

  panels = [{
    type = "esql_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    esql_control_config = {
      variable_name    = "vfq_col"
      variable_type    = "values"
      esql_query       = "ROW x = 1 | KEEP x"
      control_type     = "VALUES_FROM_QUERY"
      selected_options = []
    }
  }]
}
