variable "list_id" {
  description = "The exception list ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_exception_container" "test" {
  list_id       = var.list_id
  name          = "Agnostic Exception Container"
  description   = "Agnostic container description"
  type          = "detection"
  namespace_type = "agnostic"
}
