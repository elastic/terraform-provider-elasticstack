resource "elasticstack_kibana_security_exception_list" "example" {
  list_id        = "my-exception-list"
  name           = "My Exception List"
  description    = "List of exceptions for security rules"
  type           = "detection"
  namespace_type = "single"

  tags = ["security", "detections"]
}

resource "elasticstack_kibana_security_exception_item" "example" {
  list_id        = elasticstack_kibana_security_exception_list.example.list_id
  item_id        = "my-exception-item"
  name           = "My Exception Item"
  description    = "Exclude specific processes from alerts"
  type           = "simple"
  namespace_type = "single"

  entries = jsonencode([
    {
      field    = "process.name"
      operator = "included"
      type     = "match"
      value    = "trusted-process"
    }
  ])

  tags = ["trusted", "whitelisted"]
}
