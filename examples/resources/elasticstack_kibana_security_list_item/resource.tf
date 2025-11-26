# First create a security list
resource "elasticstack_kibana_security_list" "my_list" {
  list_id     = "allowed_domains"
  name        = "Allowed Domains"
  description = "List of allowed domains"
  type        = "keyword"
}

# Add an item to the list
resource "elasticstack_kibana_security_list_item" "domain_example" {
  list_id = elasticstack_kibana_security_list.my_list.list_id
  value   = "example.com"
}
