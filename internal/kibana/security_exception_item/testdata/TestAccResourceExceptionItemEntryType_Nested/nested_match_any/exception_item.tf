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
  name           = "Test Exception List for Nested Match Any Entry"
  description    = "Test exception list for nested match_any entry type"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Nested Match Any Entry"
  description    = "Test exception item with nested match_any entry type"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type  = "nested"
      field = "parent.field"
      entries = [
        {
          type     = "match_any"
          field    = "nested.field"
          operator = "included"
          values   = ["value1", "value2", "value3"]
        }
      ]
    }
  ]
  tags = ["test", "nested", "match_any"]
}
