variable "list_id" {}
variable "list_item_id" {}
variable "value" {}

resource "elasticstack_kibana_security_list_data_streams" "test" {
}

# First create a security list to put items in
resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = "Test List for Items with Custom ID"
  description = "A test security list for items with custom list_item_id"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

# Create a list item with custom list_item_id
resource "elasticstack_kibana_security_list_item" "test" {
  list_id      = elasticstack_kibana_security_list.test.list_id
  list_item_id = var.list_item_id
  value        = var.value

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
