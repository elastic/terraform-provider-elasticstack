variable "list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.list_id
  name           = "Test Exception List - Agnostic"
  description    = "Test exception list with agnostic namespace type"
  type           = "detection"
  namespace_type = "agnostic"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Agnostic Updated"
  description    = "Updated agnostic exception item"
  type           = "simple"
  namespace_type = "agnostic"
  entries = [
    {
      type     = "match"
      field    = "process.name"
      operator = "included"
      value    = "updated-process"
    }
  ]
  tags = ["test", "updated"]
}
