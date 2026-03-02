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
  description = "Updated test query security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  author      = ["Test Author"]
  tags        = ["test", "automation"]
  license     = "Elastic License v2"

  namespace                            = "updated-namespace"
  rule_name_override                   = "Updated Custom Query Rule Name"
  timestamp_override                   = "event.ingested"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "range" = {
        "@timestamp" = {
          "gte" = "now-1h"
          "lte" = "now"
        }
      }
    },
    {
      "terms" = {
        "event.action" = ["login", "logout", "access"]
      }
    }
  ])

  investigation_fields = ["user.name", "event.action", "source.ip"]

  risk_score_mapping = [
    {
      field      = "event.risk_level"
      operator   = "equals"
      value      = "critical"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "linux"
      version     = "2.0.0"
      integration = "auditd"
    },
    {
      package = "network"
      version = "1.5.0"
    }
  ]

  required_fields = [
    {
      name = "event.category"
      type = "keyword"
    },
    {
      name = "process.name"
      type = "keyword"
    },
    {
      name = "custom.field"
      type = "text"
    }
  ]

  severity_mapping = [
    {
      field    = "alert.severity"
      operator = "equals"
      value    = "high"
      severity = "high"
    },
    {
      field    = "alert.severity"
      operator = "equals"
      value    = "medium"
      severity = "medium"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "incident_response_pack"
        timeout = 600
        ecs_mapping = {
          "host.name"    = "hostname"
          "user.name"    = "username"
          "process.name" = "process_name"
        }
        queries = [
          {
            id       = "query1"
            query    = "SELECT * FROM logged_in_users;"
            platform = "linux"
            version  = "4.6.0"
          },
          {
            id       = "query2"
            query    = "SELECT * FROM processes WHERE state = 'R';"
            platform = "linux"
            version  = "4.6.0"
            ecs_mapping = {
              "process.pid"          = "pid"
              "process.command_line" = "cmdline"
            }
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "kill-process"
        comment = "Kill suspicious process identified during investigation"
        config = {
          field     = "process.entity_id"
          overwrite = false
        }
      }
    }
  ]
}

