variable "saved_query_id" {
  type = string
}

variable "space_id" {
  type    = string
  default = "default"
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = var.saved_query_id
  space_id       = var.space_id
  query          = "SELECT pid, name FROM processes LIMIT 10;"
  interval       = 7200
  snapshot       = true
  removed        = true
}
