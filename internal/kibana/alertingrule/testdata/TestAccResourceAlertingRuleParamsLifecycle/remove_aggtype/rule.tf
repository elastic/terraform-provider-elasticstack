variable "name" {
  description = "The rule name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name        = var.name
  rule_id     = "ff33ce2d-9fc4-5131-a350-b5bd6482799f"
  consumer    = "alerts"
  notify_when = "onActiveAlert"
  params = jsonencode({
    timeWindowSize      = 10
    timeWindowUnit      = "s"
    threshold           = [10]
    thresholdComparator = ">"
    index               = ["test-index"]
    timeField           = "@timestamp"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
}

