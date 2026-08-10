variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name
}
