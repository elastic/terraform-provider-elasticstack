provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "alerting-rule-action-freq-index"
  mappings = jsonencode({
    properties = {
      alert_date = { type = "date", format = "date_optional_time||epoch_millis" }
    }
  })
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = elasticstack_elasticsearch_index.my_index.name
    executionTimeField = "alert_date"
  })
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
