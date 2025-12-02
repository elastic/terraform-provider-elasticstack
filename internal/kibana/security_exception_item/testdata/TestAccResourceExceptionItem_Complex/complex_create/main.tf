variable "list_id" {
  type = string
}

variable "item_id" {
  type = string
}

resource "elasticstack_kibana_security_exception_list" "test" {
  name           = "test exception list for complex item"
  description    = "test exception list for complex item"
  type           = "detection"
  list_id        = var.list_id
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Complex Exception Item"
  description    = "Test complex exception item for acceptance tests"
  type           = "simple"
  namespace_type = "single"
  os_types       = ["linux", "macos"]
  tags           = ["test", "complex"]

  entries = [{
    type     = "match"
    field    = "process.name"
    operator = "included"
    value    = "test-process"
  }]
}
