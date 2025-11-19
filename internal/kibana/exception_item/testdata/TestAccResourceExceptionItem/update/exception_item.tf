variable "list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_exception_container" "container" {
  list_id     = var.list_id
  name        = "Test Exception Container"
  description = "Test container for exception item"
  type        = "detection"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id     = var.list_id
  item_id      = var.item_id
  name         = "Updated Exception Item"
  description  = "Updated item description"
  entries      = jsonencode([
    {
      field    = "source.ip"
      operator = "included"
      type     = "match"
      value    = "192.168.1.2"
    }
  ])
  tags         = ["tag1", "tag2"]
  os_types     = ["linux"]
  depends_on   = [elasticstack_kibana_security_exception_container.container]
}
