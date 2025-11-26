# First create an IP address list
resource "elasticstack_kibana_security_list" "ip_list" {
  list_id     = "allowed_ips"
  name        = "Allowed IP Addresses"
  description = "List of allowed IP addresses"
  type        = "ip"
}

# Add an IP address to the list
resource "elasticstack_kibana_security_list_item" "ip_example" {
  list_id = elasticstack_kibana_security_list.ip_list.list_id
  value   = "192.168.1.1"
}
