variable "list_id" {}
variable "value" {}

resource "elasticstack_kibana_security_list_data_streams" "test" {
}

# First create a security list to put items in
resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = "Test List for Items"
  description = "A test security list for IP addresses"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

# Create a list item
resource "elasticstack_kibana_security_list_item" "test" {
  list_id = elasticstack_kibana_security_list.test.list_id
  value   = var.value

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
