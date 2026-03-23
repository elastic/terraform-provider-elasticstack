variable "name" {
  description = "The rule name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name     = var.name
  rule_id  = "bf33ce2d-9fc4-5131-a350-b5bd6482735c"
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
  interval     = "10m"
  enabled      = false
  tags         = ["first", "second"]

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}} 1",
        "rule_name" : "{{rule.name}} 2",
        "message" : "{{context.message}} 3"
      }]
    })

    frequency {
      summary     = false
      notify_when = "onActiveAlert"
      throttle    = "2h"
    }
  }
}
