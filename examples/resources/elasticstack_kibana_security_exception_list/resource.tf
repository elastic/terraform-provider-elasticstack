resource "elasticstack_kibana_security_exception_list" "example" {
  list_id        = "my-detection-exception-list"
  name           = "My Detection Exception List"
  description    = "List of exceptions for security detection rules"
  type           = "detection"
  namespace_type = "single"

  tags = ["security", "detections"]
}
