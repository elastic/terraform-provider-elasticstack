variable "space_id" {
  type = string
}

variable "a_id" {
  type = string
}

variable "b_id" {
  type = string
}

variable "c_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_alerting_rule" "c" {
  name         = "c-rule"
  rule_id      = var.c_id
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

resource "elasticstack_kibana_alerting_rule" "a" {
  name         = "a-rule"
  rule_id      = var.a_id
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

resource "elasticstack_kibana_alerting_rule" "b" {
  name         = "b-rule"
  rule_id      = var.b_id
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
    elasticstack_kibana_alerting_rule.a,
    elasticstack_kibana_alerting_rule.b,
    elasticstack_kibana_alerting_rule.c,
  ]
}
