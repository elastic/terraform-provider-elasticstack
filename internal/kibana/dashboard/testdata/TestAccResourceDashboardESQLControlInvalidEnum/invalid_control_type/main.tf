resource "elasticstack_kibana_dashboard" "test" {
  title                  = "invalid-control-type-test"
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
    grid = { x = 0, y = 0, w = 24, h = 6 }
    esql_control_config = {
      selected_options = []
      variable_name    = "v"
      variable_type    = "values"
      esql_query       = "FROM *"
      control_type     = "UNSUPPORTED"
    }
  }]
}
