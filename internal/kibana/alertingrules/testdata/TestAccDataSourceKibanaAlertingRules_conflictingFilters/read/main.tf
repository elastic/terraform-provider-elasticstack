provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_alerting_rules" "test" {
  rule_id = "some-rule-id"
  filter  = "alert.attributes.enabled: true"
}
