variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threshold"
  query       = "event.action:(login OR logout)"
  language    = "kuery"
  enabled     = true
  description = "Updated test threshold security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  data_view_id = "updated-threshold-data-view-id"
  namespace    = "updated-threshold-namespace"

  author  = ["Test Author"]
  tags    = ["test", "threshold", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom Threshold Rule Name"
  timestamp_override                   = "event.start"
  timestamp_override_fallback_disabled = true

  filters = jsonencode([
    {
      "bool" = {
        "should" = [
          {
            "match" = {
              "user.roles" = "admin"
            }
          },
          {
            "term" = {
              "event.severity" = "high"
            }
          }
        ]
        "minimum_should_match" = 1
      }
    }
  ])

  investigation_fields = ["user.name", "source.ip", "event.outcome"]

  threshold = {
    value = 20
    field = ["user.name", "source.ip"]
  }

  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "failure"
      risk_score = 90
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
      name = "event.action"
      type = "keyword"
    },
    {
      name = "source.ip"
      type = "ip"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "failure"
      severity = "high"
    }
  ]

  alert_suppression = {
    duration = "45h"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "login_failure_investigation"
        timeout = 350
        ecs_mapping = {
          "event.outcome" = "outcome"
          "source.ip"     = "source_ip"
          "user.name"     = "username"
        }
        queries = [
          {
            id       = "failed_login_query"
            query    = "SELECT * FROM last WHERE type = 7 AND username = '{{user.name}}';"
            platform = "linux"
            version  = "4.9.0"
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to multiple failed login attempts"
      }
    }
  ]
}

