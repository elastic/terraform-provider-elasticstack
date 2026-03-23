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

variable "namespace_type" {
  type = string
}

variable "tags" {
  type = list(string)
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.list_id
  name           = var.name
  description    = var.description
  type           = var.type
  namespace_type = var.namespace_type
  tags           = var.tags
}
