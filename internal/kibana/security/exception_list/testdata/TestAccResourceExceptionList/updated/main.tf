variable "list_name" {
  type = string
}

resource "elasticstack_kibana_security_exception_list" "test" {
  name          = "${var.list_name}-updated"
  description   = "Updated exception list description"
  type          = "detection"
  namespace_type = "single"
  tags          = ["tag1", "tag2"]
}
