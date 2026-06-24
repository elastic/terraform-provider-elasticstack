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
  name     = "Osquery Saved Query Space Replace"
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = var.saved_query_id
  space_id       = elasticstack_kibana_space.test.space_id
  query          = "SELECT pid FROM processes LIMIT 1;"
  description    = "Space replace source"
}
