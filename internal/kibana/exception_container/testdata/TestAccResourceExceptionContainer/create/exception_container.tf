variable "list_id" {
  description = "The exception list ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_exception_container" "test" {
  list_id     = var.list_id
  name        = "Test Exception Container"
  description = "Test description"
  type        = "detection"
}
