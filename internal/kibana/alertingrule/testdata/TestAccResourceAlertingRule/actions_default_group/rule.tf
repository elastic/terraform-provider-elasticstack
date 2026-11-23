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

# Neither `enabled` nor `actions.group` is set, so their schema defaults apply.
resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = var.name
  rule_id      = var.rule_id
  consumer     = "monitoring"
  rule_type_id = "monitoring_alert_cluster_health"
  interval     = "1m"
  params       = jsonencode({ duration = "5m" })

  actions {
    id = elasticstack_kibana_action_connector.index_example.connector_id
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = false
      notify_when = "onActiveAlert"
    }
  }
}
