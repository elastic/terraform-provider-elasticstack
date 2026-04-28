provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name        = var.role_name
  description = "initial description"
  cluster     = ["monitor"]
  metadata = jsonencode({
    source = "terraform"
  })
  indices {
    names      = ["logs-*"]
    privileges = ["read"]
  }
}
