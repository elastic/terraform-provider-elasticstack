variable "exception_list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

variable "value_list_id" {
  description = "The value list ID"
  type        = string
}
variable "value_list_value" {
  description = "The value list value"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_list_data_streams" "test" {
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.exception_list_id
  name           = "Test Exception List for List Entry - IP"
  description    = "Test exception list for list entry type with ip"
  type           = "detection"
  namespace_type = "single"
}

# Create a value list to reference in the exception item
resource "elasticstack_kibana_security_list" "test-ip" {
  list_id     = var.value_list_id
  name        = "Test Value List - IP"
  description = "Test value list for list entry type with ip"
  type        = "ip"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]

  lifecycle {
    create_before_destroy = true
  }
}

resource "elasticstack_kibana_security_list_item" "test-item" {
  list_id = elasticstack_kibana_security_list.test-ip.list_id
  value   = var.value_list_value

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - List Entry IP"
  description    = "Test exception item with list entry type using ip"
  type           = "simple"
  namespace_type = "single"
  entries = [
    {
      type     = "list"
      field    = "source.ip"
      operator = "included"
      list = {
        id   = elasticstack_kibana_security_list.test-ip.list_id
        type = "ip"
      }
    }
  ]
  tags = ["test", "list", "ip"]
}
