resource "elasticstack_kibana_dashboard" "test" {
  title = "config-json-esql-test"
  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
  panels = [{
    type        = "esql_control"
    grid        = { x = 0, y = 0 }
    config_json = "{}"
  }]
}
