variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "esql"
  query       = "FROM logs-* | WHERE event.action == \"logout\" | STATS count(*) BY user.name, source.ip"
  language    = "esql"
  enabled     = true
  description = "Updated test ESQL security detection rule"
  severity    = "high"
  risk_score  = 80
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  author  = ["Test Author"]
  tags    = ["test", "esql", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom ESQL Rule Name"
  timestamp_override                   = "event.start"
  timestamp_override_fallback_disabled = false

  investigation_fields = ["user.name", "user.domain", "event.outcome"]

  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "failure"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "2.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "event.outcome"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "failure"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "failed_login_investigation"
        timeout        = 500
        ecs_mapping = {
          "event.outcome" = "outcome"
          "user.name"     = "username"
          "source.ip"     = "source_ip"
        }
      }
    }
  ]

  exceptions_list = [
    {
      id             = "esql-exception-1"
      list_id        = "esql-rule-exceptions"
      namespace_type = "single"
      type           = "detection"
    }
  ]
}

