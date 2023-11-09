provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "example" {
  name        = "%s"
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
