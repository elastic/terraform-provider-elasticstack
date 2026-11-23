provider "elasticstack" {
  elasticsearch {}
}

variable "role_name" {
  type = string
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name
}
