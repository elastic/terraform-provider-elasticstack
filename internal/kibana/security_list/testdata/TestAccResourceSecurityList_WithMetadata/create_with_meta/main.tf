variable "list_id" {
  type = string
}

variable "name" {
  type = string
}

variable "description" {
  type = string
}

variable "type" {
  type = string
}

variable "meta" {
  type = string
}

resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = var.name
  description = var.description
  type        = var.type
  meta        = var.meta
}
