# Create list data streams in the default space
resource "elasticstack_kibana_security_list_data_streams" "default" {
}

# Create list data streams in a custom space
resource "elasticstack_kibana_security_list_data_streams" "custom" {
  space_id = "my-space"
}
