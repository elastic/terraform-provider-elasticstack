variable "list_id" {
  description = "The exception list ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_exception_container" "test" {
  list_id     = var.list_id
  name        = "Updated Exception Container"
  description = "Updated description"
  type        = "detection"
  tags        = ["tag1", "tag2"]
  os_types    = ["linux", "windows"]
}
