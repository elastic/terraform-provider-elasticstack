variable "space_id" {}
variable "list_id" {}
variable "value" {}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Test Space for List Items"
}

resource "elasticstack_kibana_security_list_data_streams" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}

# First create a security list to put items in
resource "elasticstack_kibana_security_list" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  list_id     = var.list_id
  name        = "Test List for Items"
  description = "A test security list for IP addresses"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

# Create a list item
resource "elasticstack_kibana_security_list_item" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  list_id  = elasticstack_kibana_security_list.test.list_id
  value    = var.value

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
