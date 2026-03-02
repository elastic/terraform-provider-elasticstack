variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_data_view" "test" {
  data_view = {
    id    = "eql-data-view-id"
    title = "eql-data-view-id"
  }
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "eql"
  query       = "process where process.name == \"cmd.exe\""
  language    = "eql"
  enabled     = true
  description = "Test EQL security detection rule"
  severity    = "high"
  risk_score  = 70
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  tiebreaker_field = "@timestamp"
  data_view_id     = elasticstack_kibana_data_view.test.data_view.id
  namespace        = "eql-namespace"

  rule_name_override                   = "Custom EQL Rule Name"
  timestamp_override                   = "process.start"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "bool" = {
        "filter" = [
          {
            "term" = {
              "process.parent.name" = "explorer.exe"
            }
          }
        ]
      }
    }
  ])

  investigation_fields = ["process.name", "process.executable"]

  risk_score_mapping = [
    {
      field      = "process.executable"
      operator   = "equals"
      value      = "C:\\Windows\\System32\\cmd.exe"
      risk_score = 75
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
      name = "process.name"
      type = "keyword"
    },
    {
      name = "event.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.severity_level"
      operator = "equals"
      value    = "high"
      severity = "high"
    }
  ]

  alert_suppression = {
    group_by                = ["process.name", "user.name"]
    duration                = "10m"
    missing_fields_strategy = "suppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "suspicious_processes"
        timeout        = 300
      }
    }
  ]
}

