Manages Kibana security lists (also known as value lists). Security lists are used by exception items to define sets of values for matching or excluding in security rules.

## Example Usage

```terraform
resource "elasticstack_kibana_security_list" "ip_list" {
  space_id    = "default"
  name        = "Trusted IP Addresses"
  description = "List of trusted IP addresses for security rules"
  type        = "ip"
}

resource "elasticstack_kibana_security_list" "keyword_list" {
  space_id    = "security"
  list_id     = "custom-keywords"
  name        = "Custom Keywords"
  description = "Custom keyword list for detection rules"
  type        = "keyword"
}
```

## Notes

- Security lists define the type of data they can contain via the `type` attribute
- Once created, the `type` of a list cannot be changed
- Lists can be referenced by exception items to create more sophisticated matching rules
- The `list_id` is auto-generated if not provided
