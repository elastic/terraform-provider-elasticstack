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
  rule_id  = "af22bd1c-8fb3-4020-9249-a4ac54716255"
  consumer = "alerts"
  params = jsonencode({
    "timeSize" : 5,
    "timeUnit" : "m",
    "logView" : {
      "type" : "log-view-reference",
      "logViewId" : "default"
    },
    "count" : {
      "value" : 75,
      "comparator" : "more than"
    },
    "criteria" : [
      {
        "field" : "_id",
        "comparator" : "matches",
        "value" : "33"
      }
    ]
  })
  rule_type_id = "logs.alert.document.count"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "logs.threshold.fired"
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

    alerts_filter {
      timeframe {
        days        = [1, 2, 3]
        timezone    = "Africa/Accra"
        hours_start = "01:00"
        hours_end   = "07:00"
      }
    }
  }
}
