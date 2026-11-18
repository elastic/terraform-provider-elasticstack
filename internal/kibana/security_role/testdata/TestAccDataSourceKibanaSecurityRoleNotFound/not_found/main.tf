provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "role_name" {
  type = string
}

data "elasticstack_kibana_security_role" "test" {
  name = var.role_name
}
