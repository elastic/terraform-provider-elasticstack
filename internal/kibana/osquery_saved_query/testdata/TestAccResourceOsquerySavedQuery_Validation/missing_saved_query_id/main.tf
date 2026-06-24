provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  query = "SELECT pid FROM processes LIMIT 1;"
}
