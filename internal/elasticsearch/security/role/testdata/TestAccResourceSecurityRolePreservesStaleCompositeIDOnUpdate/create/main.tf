provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name

  cluster = ["all"]

  metadata = jsonencode({
    version = 1
  })
}
