variable "saved_query_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = var.saved_query_id
  query          = "SELECT pid FROM processes LIMIT 1;"
  interval       = 3600
  platform       = ["ios"]
}
