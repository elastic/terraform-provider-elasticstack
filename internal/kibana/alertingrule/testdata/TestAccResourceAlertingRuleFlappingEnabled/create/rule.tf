variable "name" {
  description = "The rule name"
  type        = string
}

variable "rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "${var.name}-index"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name     = var.name
  rule_id  = var.rule_id
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
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = true
      notify_when = "onActionGroupChange"
      throttle    = "10m"
    }
  }

  flapping = {
    look_back_window        = 10
    status_change_threshold = 3
    enabled                 = true
  }
}
