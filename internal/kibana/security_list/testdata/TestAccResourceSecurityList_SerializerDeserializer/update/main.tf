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

variable "serializer" {
  type = string
}

variable "deserializer" {
  type = string
}

resource "elasticstack_kibana_security_list" "test" {
  list_id      = var.list_id
  name         = var.name
  description  = var.description
  type         = var.type
  serializer   = var.serializer
  deserializer = var.deserializer
}
