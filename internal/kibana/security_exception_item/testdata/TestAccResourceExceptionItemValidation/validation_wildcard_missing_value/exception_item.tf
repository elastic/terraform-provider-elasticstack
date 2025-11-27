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
  name           = "Test Exception List"
  description    = "Test exception list for validation"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - Wildcard Missing Value"
  description    = "Test validation: wildcard entry without value"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "wildcard"
      field    = "file.path"
      operator = "included"
      # Missing value - should trigger validation error
    }
  ]
}
