resource "elasticstack_kibana_security_exception_list" "example" {
  list_id        = "my-exception-list"
  name           = "My Exception List"
  description    = "List of exceptions"
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "complex_entry" {
  list_id        = elasticstack_kibana_security_exception_list.example.list_id
  item_id        = "complex-exception"
  name           = "Complex Exception with Multiple Entries"
  description    = "Exception with multiple conditions"
  type           = "simple"
  namespace_type = "single"

  # Multiple entries with different operators
  entries = jsonencode([
    {
      field    = "host.name"
      operator = "included"
      type     = "match"
      value    = "trusted-host"
    },
    {
      field    = "user.name"
      operator = "excluded"
      type     = "match_any"
      value    = ["admin", "root"]
    }
  ])

  os_types = ["linux"]
  tags     = ["complex", "multi-condition"]
}
