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
  name           = "Test Exception List for Namespace Type Default"
  description    = "Test exception list for namespace_type default test"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id     = elasticstack_kibana_security_exception_list.test.list_id
  item_id     = var.item_id
  name        = "Test Exception Item - Namespace Type Default"
  description = "Test exception item without an explicit namespace_type"
  type        = "simple"

  entries = [{
    type     = "match"
    field    = "process.name"
    operator = "included"
    value    = "test-process"
  }]
}
