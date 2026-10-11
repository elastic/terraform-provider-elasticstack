provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_alerting_rules" "test" {}
