variable "list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

variable "name" {
  description = "The exception item name"
  type        = string
}

variable "description" {
  description = "The exception item description"
  type        = string
}

variable "type" {
  description = "The exception item type"
  type        = string
}

variable "namespace_type" {
  description = "The namespace type"
  type        = string
}

variable "tags" {
  description = "Tags for the exception item"
  type        = list(string)
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.list_id
  name           = "Test Exception List for Item"
  description    = "Test exception list"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = var.name
  description    = var.description
  type           = var.type
  namespace_type = var.namespace_type
  expire_time    = "2026-12-31T23:59:59.001Z"
  entries = [
    {
      type     = "match"
      field    = "process.name"
      operator = "included"
      value    = "test-process"
    }
  ]
  tags = var.tags
}
