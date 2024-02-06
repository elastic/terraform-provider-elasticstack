provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_data_view" "my_data_view" {
  data_view = {
    name            = "logs-*"
    title           = "logs-*"
    time_field_name = "@timestamp"
    namespaces      = ["backend"]
  }
}
