variable "space_id" {
  type = string
}

variable "rule_id" {
  type = string
}

variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_alerting_rule" "test" {
  name         = var.name
  rule_id      = var.rule_id
  space_id     = elasticstack_kibana_space.test.space_id
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
  params = jsonencode({
    aggType             = "avg"
    groupBy             = "top"
    termSize            = 10
    timeWindowSize      = 10
    timeWindowUnit      = "s"
    threshold           = [10]
    thresholdComparator = ">"
    index               = ["test-index"]
    timeField           = "@timestamp"
    aggField            = "version"
    termField           = "name"
  })
}

data "elasticstack_kibana_alerting_rules" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  filter   = "alert.attributes.name: \"this-rule-does-not-exist\""

  depends_on = [elasticstack_kibana_alerting_rule.test]
}
