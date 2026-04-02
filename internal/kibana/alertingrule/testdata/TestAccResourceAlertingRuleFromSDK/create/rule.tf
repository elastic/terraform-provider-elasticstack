variable "name" {
  type = string
}

variable "rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = var.name
  rule_id      = var.rule_id
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true

  params = jsonencode({
    "index" : [".test-index"],
    "timeField" : "@timestamp",
    "aggType" : "count",
    "groupBy" : "all",
    "timeWindowSize" : 5,
    "timeWindowUnit" : "m",
    "thresholdComparator" : ">",
    "threshold" : [1000]
  })
}
