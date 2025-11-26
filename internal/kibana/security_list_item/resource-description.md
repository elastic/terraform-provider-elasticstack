---
subcategory: "Kibana"
layout: ""
page_title: "Elasticstack: elasticstack_kibana_security_list_item Resource"
description: |-
  Manages items within Kibana security value lists.
---

# Resource: elasticstack_kibana_security_list_item

Manages items within Kibana security value lists. Value lists are containers for values that can be used within exception lists to define conditions. This resource allows you to add, update, and remove individual values (items) in those lists.

Value list items are used to store data values that match the type of their parent security list (e.g., IP addresses, keywords, etc.). These items can then be referenced in exception list entries to define exception conditions.

## Example Usage

```terraform
# First create a security list
resource "elasticstack_kibana_security_list" "ip_list" {
  list_id     = "allowed_ips"
  name        = "Allowed IP Addresses"
  description = "List of IP addresses that are allowed"
  type        = "ip"
}

# Add an IP address to the list
resource "elasticstack_kibana_security_list_item" "ip_item_1" {
  list_id = elasticstack_kibana_security_list.ip_list.list_id
  value   = "192.168.1.1"
}

# Add another IP address
resource "elasticstack_kibana_security_list_item" "ip_item_2" {
  list_id = elasticstack_kibana_security_list.ip_list.list_id
  value   = "10.0.0.1"
}

# Add a keyword item with metadata
resource "elasticstack_kibana_security_list" "keyword_list" {
  list_id     = "allowed_domains"
  name        = "Allowed Domains"
  description = "List of domains that are allowed"
  type        = "keyword"
}

resource "elasticstack_kibana_security_list_item" "domain_item" {
  list_id = elasticstack_kibana_security_list.keyword_list.list_id
  value   = "example.com"
  meta    = jsonencode({
    note = "Primary corporate domain"
  })
}
```

## Note on Space Support

**Important**: The generated Kibana API client does not currently support space_id for list item operations. While the `space_id` attribute is available in the schema for future compatibility, list items currently operate in the default space only. This is a known limitation that will be addressed in a future update when the API client is regenerated with proper space support.
