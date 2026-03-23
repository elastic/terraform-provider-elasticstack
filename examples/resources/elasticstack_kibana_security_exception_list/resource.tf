resource "elasticstack_kibana_security_exception_list" "endpoint" {
  list_id        = "my-endpoint-exception-list"
  name           = "My Endpoint Exception List"
  description    = "List of endpoint exceptions"
  type           = "endpoint"
  namespace_type = "agnostic"

  os_types = ["linux", "windows", "macos"]
  tags     = ["endpoint", "security"]
}
