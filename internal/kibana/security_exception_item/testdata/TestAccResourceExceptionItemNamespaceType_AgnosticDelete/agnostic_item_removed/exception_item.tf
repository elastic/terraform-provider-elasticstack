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

# The item is intentionally absent so Terraform deletes it on its own, while the
# owning list stays in place and cannot mask the deletion by cascading.
resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.list_id
  name           = "Test Exception List - Agnostic"
  description    = "Test exception list with agnostic namespace type"
  type           = "detection"
  namespace_type = "agnostic"
}
