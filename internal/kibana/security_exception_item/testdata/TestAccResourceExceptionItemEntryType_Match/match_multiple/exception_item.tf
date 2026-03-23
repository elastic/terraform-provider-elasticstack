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
  name           = "Test Exception List for Match Entry"
  description    = "Test exception list for match entry type"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Match Entry Multiple"
  description    = "Test exception item with multiple match entries"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "match"
      field    = "process.name"
      operator = "included"
      value    = "test-process"
    },
    {
      type     = "match"
      field    = "user.name"
      operator = "included"
      value    = "test-user"
    },
    {
      type     = "match"
      field    = "host.name"
      operator = "excluded"
      value    = "excluded-host"
    }
  ]
  tags = ["test", "match", "multiple"]
}
