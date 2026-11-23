variable "space_id" {
  type = string
}

variable "rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_alerting_rules" "test" {
  space_id = var.space_id
  rule_id  = var.rule_id
}
