variable "name" {
  description = "The rule name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule_disabled" {
  name        = var.name
  rule_id     = "df33ce2d-9fc4-5131-a350-b5bd6482737d"
  consumer    = "alerts"
  notify_when = "onActiveAlert"
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
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = false
}
