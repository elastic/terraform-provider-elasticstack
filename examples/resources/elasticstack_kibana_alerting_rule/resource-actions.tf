provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_kibana_alerting_rule" "example" {
  name         = "%s"
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

  actions {
    id    = elasticstack_kibana_action_connector.example.connector_type_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "alert_id" : "{{alert.id}}"
      }]
    })
  }
}
