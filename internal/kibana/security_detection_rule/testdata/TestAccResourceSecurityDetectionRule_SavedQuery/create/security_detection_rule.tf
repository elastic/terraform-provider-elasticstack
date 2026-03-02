variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "saved_query"
  query       = "*:*"
  enabled     = true
  description = "Test saved query security detection rule"
  severity    = "low"
  risk_score  = 30
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  saved_id     = "test-saved-query-id"
  data_view_id = "saved-query-data-view-id"
  namespace    = "saved-query-namespace"

  rule_name_override                   = "Custom Saved Query Rule Name"
  timestamp_override                   = "event.start"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "prefix" = {
        "event.action" = "user_"
      }
    }
  ])

  investigation_fields = ["event.category", "event.action"]

  risk_score_mapping = [
    {
      field      = "event.category"
      operator   = "equals"
      value      = "authentication"
      risk_score = 45
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "1.0.0"
      integration = "logs"
    }
  ]

  required_fields = [
    {
      name = "event.category"
      type = "keyword"
    },
    {
      name = "event.action"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.category"
      operator = "equals"
      value    = "authentication"
      severity = "low"
    }
  ]

  alert_suppression = {
    group_by                = ["event.category", "event.action"]
    duration                = "8h"
    missing_fields_strategy = "suppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM logged_in_users WHERE user = '{{user.name}}';"
        timeout = 250
        ecs_mapping = {
          "event.category" = "category"
          "event.action"   = "action"
          "user.name"      = "username"
        }
      }
    }
  ]
}

