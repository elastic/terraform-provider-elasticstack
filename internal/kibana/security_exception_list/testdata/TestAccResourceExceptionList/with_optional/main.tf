resource "elasticstack_kibana_security_exception_list" "test" {
  name           = "Test Exception List"
  description    = "This is a test exception list"
  type           = "endpoint"
  namespace_type = "single"
  os_types       = ["linux", "windows"]
  tags           = ["test", "demo"]
}