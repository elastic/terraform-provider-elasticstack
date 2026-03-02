variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Test query security detection rule"
  severity    = "medium"
  risk_score  = 50
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  namespace   = "test-namespace"

  rule_name_override                   = "Custom Query Rule Name"
  timestamp_override                   = "@timestamp"
  timestamp_override_fallback_disabled = true

  filters = jsonencode([
    {
      "bool" = {
        "must" = [
          {
            "term" = {
              "event.category" = "authentication"
            }
          }
        ]
        "must_not" = [
          {
            "term" = {
              "event.outcome" = "success"
            }
          }
        ]
      }
    }
  ])

  investigation_fields = ["user.name", "event.action"]

  risk_score_mapping = [
    {
      field      = "event.severity"
      operator   = "equals"
      value      = "high"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "windows"
      version     = "1.0.0"
      integration = "system"
    }
  ]

  required_fields = [
    {
      name = "event.type"
      type = "keyword"
    },
    {
      name = "host.os.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.severity_level"
      operator = "equals"
      value    = "critical"
      severity = "critical"
    }
  ]

  alert_suppression = {
    group_by                = ["user.name", "host.name"]
    duration                = "5m"
    missing_fields_strategy = "suppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM processes WHERE name = 'malicious.exe';"
        timeout = 300
        ecs_mapping = {
          "process.name" = "name"
          "process.pid"  = "pid"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to suspicious activity"
      }
    }
  ]
}

