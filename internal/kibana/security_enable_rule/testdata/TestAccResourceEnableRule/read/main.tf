provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id = "default"
  key      = "test_tag"
  value    = "terraform_test"
}
