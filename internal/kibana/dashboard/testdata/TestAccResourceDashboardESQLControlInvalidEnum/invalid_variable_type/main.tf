resource "elasticstack_kibana_dashboard" "test" {
  title                  = "invalid-enum-test"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kql"
  query_text             = ""
  panels = [{
    type = "esql_control"
    grid = { x = 0, y = 0, w = 24, h = 6 }
    esql_control_config = {
      selected_options = []
      variable_name    = "v"
      variable_type    = "unsupported_type"
      esql_query       = "FROM *"
      control_type     = "STATIC_VALUES"
    }
  }]
}
