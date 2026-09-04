variable "name" {
  description = "The role mapping name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = var.name
  enabled = true
  roles   = ["admin"]

  rules = jsonencode({
    field = { groups = ["project1"] }
  })
}

data "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = elasticstack_elasticsearch_security_role_mapping.test.name
}
