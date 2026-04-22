provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_enable_rule" "test" {
  key   = "test_tag"
  value = "terraform_test_default_space"
}
