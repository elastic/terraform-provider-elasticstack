variable "space_id" {
  type = string
}

variable "key" {
  type = string
}

variable "value" {
  type = string
}

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id = var.space_id
  key      = var.key
  value    = var.value
}
