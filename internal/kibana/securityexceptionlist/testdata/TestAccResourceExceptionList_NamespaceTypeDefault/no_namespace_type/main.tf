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

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id     = var.list_id
  name        = var.name
  description = var.description
  type        = var.type
}
