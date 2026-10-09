variable "rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_alerting_rules" "test" {
  rule_id = var.rule_id
}
