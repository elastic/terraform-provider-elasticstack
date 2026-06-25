provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_osquery_saved_query" "test" {}
