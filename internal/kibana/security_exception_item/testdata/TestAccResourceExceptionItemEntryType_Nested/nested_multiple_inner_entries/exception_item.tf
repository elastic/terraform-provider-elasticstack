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
  name           = "Test Exception List for Nested Entry with Multiple Inner Entries"
  description    = "Test exception list for nested entry type with multiple inner entries"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Nested Entry with Multiple Inner Entries"
  description    = "Test exception item with nested entry containing multiple inner entries"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type  = "nested"
      field = "parent.field"
      entries = [
        {
          type     = "match"
          field    = "nested.field1"
          operator = "included"
          value    = "nested-value-1"
        },
        {
          type     = "match_any"
          field    = "nested.field2"
          operator = "included"
          values   = ["value-a", "value-b", "value-c"]
        },
        {
          type     = "exists"
          field    = "nested.field3"
          operator = "included"
        }
      ]
    }
  ]
  tags = ["test", "nested", "multiple-inner"]
}
