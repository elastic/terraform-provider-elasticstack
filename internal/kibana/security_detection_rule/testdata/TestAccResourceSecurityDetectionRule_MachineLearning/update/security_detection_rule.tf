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
  description             = "Updated test ML security detection rule"
  severity                = "high"
  risk_score              = 85
  from                    = "now-6m"
  to                      = "now"
  interval                = "5m"
  anomaly_threshold       = 80
  machine_learning_job_id = ["test-ml-job", "test-ml-job-2"]

  author  = ["Test Author"]
  tags    = ["test", "ml", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom ML Rule Name"
  timestamp_override                   = "ml.anomaly_score"
  timestamp_override_fallback_disabled = true

  investigation_fields = ["ml.anomaly_score", "ml.job_id", "ml.is_anomaly"]

  risk_score_mapping = [
    {
      field      = "ml.is_anomaly"
      operator   = "equals"
      value      = "true"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "ml"
      version     = "2.0.0"
      integration = "anomaly_detection"
    }
  ]

  required_fields = [
    {
      name = "ml.is_anomaly"
      type = "boolean"
    },
    {
      name = "ml.job_id"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "ml.is_anomaly"
      operator = "equals"
      value    = "true"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "ml_anomaly_investigation"
        timeout = 700
        ecs_mapping = {
          "ml.job_id"     = "job_id"
          "ml.is_anomaly" = "is_anomaly"
          "host.name"     = "hostname"
        }
        queries = [
          {
            id       = "ml_query1"
            query    = "SELECT * FROM system_info;"
            platform = "linux"
            version  = "4.7.0"
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Collect process tree for ML anomaly investigation"
      }
    }
  ]

  exceptions_list = [
    {
      id             = "ml-exception-1"
      list_id        = "ml-rule-exceptions"
      namespace_type = "agnostic"
      type           = "detection"
    }
  ]
}

