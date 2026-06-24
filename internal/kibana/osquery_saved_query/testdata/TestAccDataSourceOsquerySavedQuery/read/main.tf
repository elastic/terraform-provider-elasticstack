variable "saved_query_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = var.saved_query_id
  query          = "SELECT pid, name FROM processes LIMIT 5;"
  description    = "Data source read fixture"
  interval       = 7200
  version        = "2.0.0"
}

data "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = elasticstack_kibana_osquery_saved_query.test.saved_query_id
}
