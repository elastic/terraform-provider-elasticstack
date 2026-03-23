variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = var.name
  rule_id      = "ef33ce2d-9fc4-5131-a350-b5bd6482745e"
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
