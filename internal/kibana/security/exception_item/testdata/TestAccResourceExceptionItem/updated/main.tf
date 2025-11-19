variable "list_name" {
  type = string
}

variable "item_name" {
  type = string
}

resource "elasticstack_kibana_security_exception_list" "test" {
  name          = var.list_name
  description   = "Test exception list for items"
  type          = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id       = elasticstack_kibana_security_exception_list.test.list_id
  name          = "${var.item_name}-updated"
  description   = "Updated exception item description"
  type          = "simple"
  namespace_type = "single"
  
  entries = jsonencode([
    {
      field    = "process.name"
      operator = "included"
      type     = "match"
      value    = "updated_process"
    }
  ])
  
  tags = ["tag1", "tag2"]
}
