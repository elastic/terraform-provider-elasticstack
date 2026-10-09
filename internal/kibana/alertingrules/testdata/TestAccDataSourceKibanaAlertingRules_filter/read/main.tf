variable "space_id" {
  type = string
}

variable "enabled_rule_id" {
  type = string
}

variable "disabled_rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_alerting_rule" "enabled" {
  name         = "enabled-rule"
  rule_id      = var.enabled_rule_id
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

resource "elasticstack_kibana_alerting_rule" "disabled" {
  name         = "disabled-rule"
  rule_id      = var.disabled_rule_id
  space_id     = elasticstack_kibana_space.test.space_id
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = false
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
  filter   = "alert.attributes.enabled: true"

  depends_on = [
    elasticstack_kibana_alerting_rule.enabled,
    elasticstack_kibana_alerting_rule.disabled,
  ]
}
