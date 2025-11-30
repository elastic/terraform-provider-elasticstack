variable "list_id" {}
variable "value" {}
variable "meta" {}

# First create a security list to put items in
resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = "Test List for Items with Meta"
  description = "A test security list for items with metadata"
  type        = "keyword"
}

# Update list item with different meta
resource "elasticstack_kibana_security_list_item" "test" {
  list_id = elasticstack_kibana_security_list.test.list_id
  value   = var.value
  meta    = var.meta
}
