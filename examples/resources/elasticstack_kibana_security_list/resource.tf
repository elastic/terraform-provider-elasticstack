resource "elasticstack_kibana_security_list" "ip_list" {
  space_id    = "default"
  name        = "Trusted IP Addresses"
  description = "List of trusted IP addresses for security rules"
  type        = "ip"
}
