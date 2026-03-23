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
  name           = "Test Exception List for Exists Entry"
  description    = "Test exception list for exists entry type"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Exists Entry Multiple"
  description    = "Test exception item with multiple exists entries"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "exists"
      field    = "file.hash.sha256"
      operator = "included"
    },
    {
      type     = "exists"
      field    = "process.code_signature.trusted"
      operator = "included"
    },
    {
      type     = "exists"
      field    = "network.protocol"
      operator = "excluded"
    }
  ]
  tags = ["test", "exists", "multiple"]
}
