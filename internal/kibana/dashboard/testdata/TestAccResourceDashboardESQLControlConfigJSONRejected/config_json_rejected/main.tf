resource "elasticstack_kibana_dashboard" "test" {
  title     = "config-json-esql-test"
  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kql"
  query_text     = ""

  panels = [{
    type        = "esql_control"
    grid        = { x = 0, y = 0 }
    config_json = "{}"
  }]
}
