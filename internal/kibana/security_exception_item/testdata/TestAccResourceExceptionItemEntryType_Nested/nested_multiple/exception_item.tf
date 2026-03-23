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
  name           = "Test Exception List for Nested Entry"
  description    = "Test exception list for nested entry type"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Nested Entry Multiple"
  description    = "Test exception item with multiple nested entries"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type  = "nested"
      field = "parent.field"
      entries = [
        {
          type     = "match"
          field    = "nested.field"
          operator = "included"
          value    = "nested-value"
        }
      ]
    },
    {
      type  = "nested"
      field = "process.parent"
      entries = [
        {
          type     = "match"
          field    = "process.parent.name"
          operator = "included"
          value    = "parent-process"
        }
      ]
    },
    {
      type  = "nested"
      field = "file.attributes"
      entries = [
        {
          type     = "exists"
          field    = "file.attributes.hidden"
          operator = "included"
        }
      ]
    }
  ]
  tags = ["test", "nested", "multiple"]
}
