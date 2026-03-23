resource "elasticstack_kibana_security_list" "keyword_list" {
  space_id    = "security"
  list_id     = "custom-keywords"
  name        = "Custom Keywords"
  description = "Custom keyword list for detection rules"
  type        = "keyword"
}
