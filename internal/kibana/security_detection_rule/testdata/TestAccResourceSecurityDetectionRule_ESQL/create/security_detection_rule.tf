variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "esql"
  query       = "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"
  language    = "esql"
  enabled     = true
  description = "Test ESQL security detection rule"
  severity    = "medium"
  risk_score  = 60
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  namespace                            = "esql-namespace"
  rule_name_override                   = "Custom ESQL Rule Name"
  timestamp_override                   = "event.created"
  timestamp_override_fallback_disabled = true

  investigation_fields = ["user.name", "user.domain"]

  risk_score_mapping = [
    {
      field      = "user.domain"
      operator   = "equals"
      value      = "admin"
      risk_score = 80
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "1.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "event.action"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.domain"
      operator = "equals"
      value    = "admin"
      severity = "high"
    }
  ]

  alert_suppression = {
    group_by                = ["user.name", "user.domain"]
    duration                = "15m"
    missing_fields_strategy = "doNotSuppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM users WHERE username LIKE '%admin%';"
        timeout = 400
        ecs_mapping = {
          "user.name"   = "username"
          "user.domain" = "domain"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to suspicious admin activity"
      }
    }
  ]
}

