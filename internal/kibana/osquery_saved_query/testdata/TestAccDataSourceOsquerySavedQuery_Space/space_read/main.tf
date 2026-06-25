variable "space_id" {
  type = string
}

variable "saved_query_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Osquery DS Space"
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = var.saved_query_id
  space_id       = elasticstack_kibana_space.test.space_id
  query          = "SELECT pid FROM processes LIMIT 1;"
  interval       = 3600
  depends_on     = [elasticstack_kibana_space.test]
}

data "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = elasticstack_kibana_osquery_saved_query.test.saved_query_id
  space_id       = elasticstack_kibana_space.test.space_id
}
