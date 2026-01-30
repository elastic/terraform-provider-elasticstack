variable "name" {
  description = "The rule name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name        = var.name
  rule_id     = "af22bd1c-8fb3-4020-9249-a4ac5471624d"
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
  enabled      = true
}
