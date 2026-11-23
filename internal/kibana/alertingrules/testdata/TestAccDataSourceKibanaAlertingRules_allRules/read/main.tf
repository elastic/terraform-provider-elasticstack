variable "space_id" {
  type = string
}

variable "first_id" {
  type = string
}

variable "second_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_alerting_rule" "first" {
  name         = "first-rule"
  rule_id      = var.first_id
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

resource "elasticstack_kibana_alerting_rule" "second" {
  name         = "second-rule"
  rule_id      = var.second_id
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

  depends_on = [
    elasticstack_kibana_alerting_rule.first,
    elasticstack_kibana_alerting_rule.second,
  ]
}
