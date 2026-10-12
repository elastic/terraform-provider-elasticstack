variable "name" {
  type = string
}

variable "tagged_rule_id" {
  type = string
}

variable "untagged_rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "tagged" {
  name         = "${var.name}-tagged"
  rule_id      = var.tagged_rule_id
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
  tags         = ["tag-one", "tag-two"]
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

resource "elasticstack_kibana_alerting_rule" "untagged" {
  name         = "${var.name}-untagged"
  rule_id      = var.untagged_rule_id
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

data "elasticstack_kibana_alerting_rules" "tagged" {
  rule_id = elasticstack_kibana_alerting_rule.tagged.rule_id
}

data "elasticstack_kibana_alerting_rules" "untagged" {
  rule_id = elasticstack_kibana_alerting_rule.untagged.rule_id
}
