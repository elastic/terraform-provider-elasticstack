variable "exception_list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

variable "value_list_id" {
  description = "The value list ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.exception_list_id
  name           = "Test Exception List for List Entry"
  description    = "Test exception list for list entry type"
  type           = "detection"
  namespace_type = "single"
}

# Create a value list to reference in the exception item
resource "elasticstack_kibana_security_value_list" "test" {
  list_id     = var.value_list_id
  name        = "Test Value List"
  description = "Test value list for list entry type"
  type        = "ip"
  values      = ["192.168.1.1", "192.168.1.2", "10.0.0.1"]
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - List Entry"
  description    = "Test exception item with list entry type"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "list"
      field    = "source.ip"
      operator = "included"
      list = {
        id   = elasticstack_kibana_security_value_list.test.list_id
        type = "ip"
      }
    }
  ]
  tags = ["test", "list"]
}
