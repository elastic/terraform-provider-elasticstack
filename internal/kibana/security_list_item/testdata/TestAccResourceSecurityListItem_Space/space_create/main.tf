variable "space_id" {}
variable "list_id" {}
variable "value" {}

# Create a dedicated space for security lists
resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Security Lists Space"
  description = "A test space for security lists and list items"
}

# Create a security list in the space
resource "elasticstack_kibana_security_list" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  list_id     = var.list_id
  name        = "IP Blocklist"
  description = "A test security list for blocking IP addresses"
  type        = "ip"
}

# Create a list item in the space
resource "elasticstack_kibana_security_list_item" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  list_id  = elasticstack_kibana_security_list.test.list_id
  value    = var.value
}
