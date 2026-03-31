resource "elasticstack_kibana_dashboard" "test" {
  title = "invalid-control-type-test"
  panels = [{
    type = "esql_control"
    grid = { x = 0, y = 0 }
    esql_control_config = {
      selected_options = []
      variable_name    = "v"
      variable_type    = "values"
      esql_query       = "FROM *"
      control_type     = "UNSUPPORTED"
    }
  }]
}
