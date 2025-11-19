variable "list_name" {
  type = string
}

resource "elasticstack_kibana_security_exception_list" "test" {
  name           = var.list_name
  description    = "Test exception list"
  type           = "detection"
  namespace_type = "single"
}
