provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "data_analyst" {
  name = "data_analyst"

  cluster = ["monitor"]

  indices {
    names      = ["logs-*"]
    privileges = ["read", "view_index_metadata"]
  }
}
