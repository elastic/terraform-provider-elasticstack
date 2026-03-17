variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                    = var.name
  type                    = "machine_learning"
  enabled                 = true
  description             = "Test ML security detection rule"
  severity                = "critical"
  risk_score              = 90
  from                    = "now-6m"
  to                      = "now"
  interval                = "5m"
  anomaly_threshold       = 75
  machine_learning_job_id = ["test-ml-job"]

  namespace                            = "ml-namespace"
  rule_name_override                   = "Custom ML Rule Name"
  timestamp_override                   = "ml.job_id"
  timestamp_override_fallback_disabled = false

  investigation_fields = ["ml.anomaly_score", "ml.job_id"]

  risk_score_mapping = [
    {
      field      = "ml.anomaly_score"
      operator   = "equals"
      value      = "critical"
      risk_score = 100
    }
  ]

  related_integrations = [
    {
      package     = "ml"
      version     = "1.0.0"
      integration = "anomaly_detection"
    }
  ]

  required_fields = [
    {
      name = "ml.anomaly_score"
      type = "double"
    },
    {
      name = "ml.job_id"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "ml.anomaly_score"
      operator = "equals"
      value    = "critical"
      severity = "critical"
    }
  ]

  alert_suppression = {
    group_by                = ["ml.job_id"]
    duration                = "30m"
    missing_fields_strategy = "suppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM processes WHERE pid IN (SELECT DISTINCT pid FROM connections WHERE remote_address NOT LIKE '10.%' AND remote_address NOT LIKE '192.168.%' AND remote_address NOT LIKE '127.%');"
        timeout = 600
        ecs_mapping = {
          "process.pid"      = "pid"
          "process.name"     = "name"
          "ml.anomaly_score" = "anomaly_score"
        }
      }
    }
  ]
}

