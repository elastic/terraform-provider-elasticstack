provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "example_with_action_frequency" {
  name     = "%s"
  consumer = "alerts"
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

  actions {
    # Should be the id of a MS Teams connector
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "message" : "foobar"
    })

    frequency {
      summary     = false
      notify_when = "onActionGroupChange"
    }
  }
}
