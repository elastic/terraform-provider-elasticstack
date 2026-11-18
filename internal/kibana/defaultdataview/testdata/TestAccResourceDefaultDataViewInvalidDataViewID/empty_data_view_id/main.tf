provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_default_data_view" "test" {
  data_view_id = ""
}
