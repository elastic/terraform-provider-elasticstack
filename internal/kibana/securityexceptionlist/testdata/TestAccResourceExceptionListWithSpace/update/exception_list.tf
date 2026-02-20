variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

variable "list_id" {
  description = "The exception list ID"
  type        = string
}

variable "name" {
  description = "The exception list name"
  type        = string
}

variable "description" {
  description = "The exception list description"
  type        = string
}

variable "type" {
  description = "The exception list type"
  type        = string
}

variable "namespace_type" {
  description = "The namespace type"
  type        = string
}

variable "tags" {
  description = "Tags for the exception list"
  type        = list(string)
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Exception Lists"
  description = "Space for testing exception lists"
}

resource "elasticstack_kibana_security_exception_list" "test" {
  space_id       = elasticstack_kibana_space.test.space_id
  list_id        = var.list_id
  name           = var.name
  description    = var.description
  type           = var.type
  namespace_type = var.namespace_type
  tags           = var.tags
}
