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
  name           = "Test Exception List for Match Any Entry"
  description    = "Test exception list for match_any entry type"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Match Any Entry"
  description    = "Test exception item with match_any entry type"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "match_any"
      field    = "process.name"
      operator = "included"
      values   = ["process1", "process2", "process3"]
    }
  ]
  tags = ["test", "match_any"]
}
