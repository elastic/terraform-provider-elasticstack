# First create a security list
resource "elasticstack_kibana_security_list" "tagged_domains" {
  list_id     = "tagged_domains"
  name        = "Tagged Domains"
  description = "Domains with associated metadata"
  type        = "keyword"
}

# Add an item with metadata
resource "elasticstack_kibana_security_list_item" "domain_with_meta" {
  list_id = elasticstack_kibana_security_list.tagged_domains.list_id
  value   = "internal.example.com"
  meta = jsonencode({
    category = "internal"
    owner    = "infrastructure-team"
    note     = "Primary internal domain"
  })
}
