variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "saved_query"
  query       = "event.action:*"
  enabled     = true
  description = "Updated test saved query security detection rule"
  severity    = "medium"
  risk_score  = 60
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  saved_id     = "test-saved-query-id-updated"
  data_view_id = "updated-saved-query-data-view-id"
  namespace    = "updated-saved-query-namespace"

  author  = ["Test Author"]
  tags    = ["test", "saved-query", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom Saved Query Rule Name"
  timestamp_override                   = "event.end"
  timestamp_override_fallback_disabled = true

  filters = jsonencode([
    {
      "script" = {
        "script" = {
          "source" = "doc['event.severity'].value > 2"
        }
      }
    }
  ])

  investigation_fields = ["host.name", "user.name", "process.name"]

  risk_score_mapping = [
    {
      field      = "event.type"
      operator   = "equals"
      value      = "access"
      risk_score = 70
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "2.0.0"
      integration = "logs"
    }
  ]

  required_fields = [
    {
      name = "event.type"
      type = "keyword"
    },
    {
      name = "host.name"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.type"
      operator = "equals"
      value    = "access"
      severity = "medium"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "access_investigation_pack"
        timeout = 400
        ecs_mapping = {
          "event.type" = "type"
          "host.name"  = "hostname"
          "user.name"  = "username"
        }
        queries = [
          {
            id       = "access_query1"
            query    = "SELECT * FROM users WHERE username = '{{user.name}}';"
            platform = "linux"
            version  = "4.8.0"
            ecs_mapping = {
              "user.id" = "uid"
            }
          }
        ]
      }
    }
  ]

  exceptions_list = [
    {
      id             = "saved-query-exception-1"
      list_id        = "saved-query-exceptions"
      namespace_type = "agnostic"
      type           = "detection"
    }
  ]
}

