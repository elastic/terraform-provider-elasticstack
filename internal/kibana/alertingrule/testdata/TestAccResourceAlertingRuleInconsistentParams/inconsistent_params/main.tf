locals {
  kafka_error_rate_threshold    = 15  # Percent
  kafka_error_rate_minimum_logs = 100 # Minimum logs to consider for the alert
  motel_logs                    = "logs-*"
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "server_log" {
  name              = "acc_test_server_log_connector"
  connector_type_id = ".server-log"
}

resource "elasticstack_kibana_alerting_rule" "kafka_error_alert" {
  name     = "[Motel Services] Kafka Error Rate"
  consumer = "infrastructure"
  tags     = ["kafka-error-rate"]
  params = jsonencode({
    searchType          = "esqlQuery",
    timeWindowSize      = 5,
    timeWindowUnit      = "m",
    threshold           = [0],
    thresholdComparator = ">",
    size                = 100,
    esqlQuery = {
      esql = <<EOF
FROM ${local.motel_logs}
  | WHERE otelcol.component.id == "kafka"
  | STATS errors = COUNT(error.message) WHERE (log.level == "error" OR log.level == "warn") AND NOT (kubernetes.namespace == "motel-ingest-collector" AND MATCH_PHRASE(error.message, "context canceled")), logs = COUNT(*) BY cluster = orchestrator.cluster.name, service = kubernetes.namespace
  | EVAL error_rate = ROUND(errors::DOUBLE / logs::DOUBLE * 100, 2)
  | WHERE error_rate::DOUBLE > ${local.kafka_error_rate_threshold} AND logs > ${local.kafka_error_rate_minimum_logs}
  | DROP errors, logs
  | SORT error_rate DESC
EOF
    },
    aggType      = "count",
    groupBy      = "all",
    sourceFields = [],
    timeField    = "@timestamp",
  })
  rule_type_id = ".es-query"
  interval     = "1m"
  enabled      = true
  alert_delay  = 2

  actions {
    group = "query matched"
    id    = elasticstack_kibana_action_connector.server_log.connector_id
    params = jsonencode({
      message = "{{rule.name}} matched"
    })
    frequency {
      notify_when = "onActionGroupChange"
      throttle    = null
      summary     = false
    }
  }
  actions {
    group = "recovered"
    id    = elasticstack_kibana_action_connector.server_log.connector_id
    params = jsonencode({
      message = "{{rule.name}} recovered"
    })
    frequency {
      notify_when = "onActionGroupChange"
      throttle    = null
      summary     = false
    }
  }
}