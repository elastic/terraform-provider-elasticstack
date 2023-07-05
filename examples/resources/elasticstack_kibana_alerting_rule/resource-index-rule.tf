provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_kibana_alerting_rule" "my_rule" {
  name         = "my_rule"
  consumer     = "alerts"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
  notify_when  = "onActiveAlert"

  params = jsonencode({
    aggType             = "count"
    thresholdComparator = ">"
    timeWindowSize      = 1
    timeWindowUnit      = "d"
    groupBy             = "all"
    threshold           = [10]
    index               = elasticstack_elasticsearch_data_stream.my_data_stream.name
    timeField           = "@timestamp"
  })

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
      "rule_id" : "{{rule.id}}",
      "rule_name": "{{rule.name}}",
      "message": "{{context.message}}"
      }]
    })
  }
}