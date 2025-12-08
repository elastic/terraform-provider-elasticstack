variable "list_id" {}
variable "value" {}
variable "meta" {}

resource "elasticstack_kibana_security_list_data_streams" "test" {
}

# First create a security list to put items in
resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = "Test List for Items with Meta"
  description = "A test security list for items with metadata"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

# Create a list item with meta
resource "elasticstack_kibana_security_list_item" "test" {
  list_id = elasticstack_kibana_security_list.test.list_id
  value   = var.value
  meta    = var.meta

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
