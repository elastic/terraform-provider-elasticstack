variable "list_id" {
  type = string
}

variable "item_id_1" {
  type = string
}

variable "item_id_2" {
  type = string
}

variable "item_id_3" {
  type = string
}

resource "elasticstack_kibana_security_exception_list" "test" {
  name           = "test exception list for multiple items"
  description    = "test exception list with multiple items"
  type           = "detection"
  list_id        = var.list_id
  namespace_type = "single"
}

# First exception item - simple match entry
resource "elasticstack_kibana_security_exception_item" "test1" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id_1
  name           = "Test Exception Item 1"
  description    = "First exception item in the list"
  type           = "simple"
  namespace_type = "single"
  os_types       = ["linux"]
  tags           = ["test", "item1"]

  entries = [{
    type     = "match"
    field    = "process.name"
    operator = "included"
    value    = "process1"
  }]
}

# Second exception item - match_any entry
resource "elasticstack_kibana_security_exception_item" "test2" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id_2
  name           = "Test Exception Item 2"
  description    = "Second exception item in the list"
  type           = "simple"
  namespace_type = "single"
  os_types       = ["linux", "macos"]
  tags           = ["test", "item2"]

  entries = [{
    type     = "match_any"
    field    = "user.name"
    operator = "included"
    values   = ["user1", "user2"]
  }]

}

# Third exception item - multiple entries (wildcard and exists)
resource "elasticstack_kibana_security_exception_item" "test3" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id_3
  name           = "Test Exception Item 3"
  description    = "Third exception item in the list"
  type           = "simple"
  namespace_type = "single"
  os_types       = ["linux", "macos", "windows"]
  tags           = ["test", "item3"]

  entries = [
    {
      type     = "wildcard"
      field    = "file.path"
      operator = "included"
      value    = "/tmp/*.tmp"
    },
    {
      type     = "exists"
      field    = "file.hash.sha256"
      operator = "included"
    }
  ]

}
