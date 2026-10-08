provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "example" {
  name         = "example-rule"
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

data "elasticstack_kibana_alerting_rules" "example" {
  rule_id = elasticstack_kibana_alerting_rule.example.rule_id
}

output "last_execution_status" {
  value = data.elasticstack_kibana_alerting_rules.example.rules[0].last_execution_status
}
