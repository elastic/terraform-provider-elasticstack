variable "space_id" {
  description = "The space ID"
  type        = string
}

variable "exception_list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

variable "value_list_id_ip" {
  description = "The value list ID for IP"
  type        = string
}

variable "value_list_id_keyword" {
  description = "The value list ID for keyword"
  type        = string
}

variable "value_list_value_ip" {
  description = "The value list value for IP"
  type        = string
}

variable "value_list_value_keyword" {
  description = "The value list value for keyword"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Test Space for List Entry"
}

resource "elasticstack_kibana_security_list_data_streams" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}

resource "elasticstack_kibana_security_exception_list" "test" {
  space_id       = elasticstack_kibana_space.test.space_id
  list_id        = var.exception_list_id
  name           = "Test Exception List for List Entry - Multiple"
  description    = "Test exception list for list entry type with multiple lists"
  type           = "detection"
  namespace_type = "single"
}

# Create IP value list
resource "elasticstack_kibana_security_list" "test-ip" {
  space_id    = elasticstack_kibana_space.test.space_id
  list_id     = var.value_list_id_ip
  name        = "Test Value List - IP"
  description = "Test value list for list entry type with ip"
  type        = "ip"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]

  lifecycle {
    create_before_destroy = true
  }
}

resource "elasticstack_kibana_security_list_item" "test-ip-item" {
  space_id = elasticstack_kibana_space.test.space_id
  list_id  = elasticstack_kibana_security_list.test-ip.list_id
  value    = var.value_list_value_ip

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

# Create keyword value list
resource "elasticstack_kibana_security_list" "test-keyword" {
  space_id    = elasticstack_kibana_space.test.space_id
  list_id     = var.value_list_id_keyword
  name        = "Test Value List - Keyword"
  description = "Test value list for list entry type with keyword"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]

  lifecycle {
    create_before_destroy = true
  }
}

resource "elasticstack_kibana_security_list_item" "test-keyword-item" {
  space_id = elasticstack_kibana_space.test.space_id
  list_id  = elasticstack_kibana_security_list.test-keyword.list_id
  value    = var.value_list_value_keyword

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

resource "elasticstack_kibana_security_exception_item" "test" {
  space_id       = elasticstack_kibana_space.test.space_id
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Test Exception Item - List Entry Multiple"
  description    = "Test exception item with multiple list entries"
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
    },
    {
      type     = "list"
      field    = "process.name"
      operator = "included"
      list = {
        id   = elasticstack_kibana_security_list.test-keyword.list_id
        type = "keyword"
      }
    },
    {
      type     = "list"
      field    = "destination.ip"
      operator = "excluded"
      list = {
        id   = elasticstack_kibana_security_list.test-ip.list_id
        type = "ip"
      }
    }
  ]
  tags = ["test", "list", "multiple"]
}
