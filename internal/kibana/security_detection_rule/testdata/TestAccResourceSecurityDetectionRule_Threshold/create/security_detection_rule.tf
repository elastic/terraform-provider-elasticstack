variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threshold"
  query       = "event.action:login"
  language    = "kuery"
  enabled     = true
  description = "Test threshold security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  data_view_id = "threshold-data-view-id"
  namespace    = "threshold-namespace"

  rule_name_override                   = "Custom Threshold Rule Name"
  timestamp_override                   = "event.created"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "bool" = {
        "filter" = [
          {
            "range" = {
              "event.ingested" = {
                "gte" = "now-24h"
              }
            }
          }
        ]
      }
    }
  ])

  investigation_fields = ["user.name", "event.action"]

  threshold = {
    value = 10
    field = ["user.name"]
  }

  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "success"
      risk_score = 45
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
      name = "event.action"
      type = "keyword"
    },
    {
      name = "user.name"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "success"
      severity = "medium"
    }
  ]

  alert_suppression = {
    duration = "30m"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM logged_in_users WHERE user = '{{user.name}}' ORDER BY time DESC LIMIT 10;"
        timeout = 200
        ecs_mapping = {
          "user.name"     = "username"
          "event.action"  = "action"
          "event.outcome" = "outcome"
        }
      }
    }
  ]
}

