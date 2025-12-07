provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name

  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  metadata = jsonencode({
    version = 1
  })
}
