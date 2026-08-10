variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = []
  run_as  = []
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
