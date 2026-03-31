resource "elasticstack_kibana_dashboard" "test" {
  title = "config-json-esql-test"
  panels = [{
    type        = "esql_control"
    grid        = { x = 0, y = 0 }
    config_json = "{}"
  }]
}
