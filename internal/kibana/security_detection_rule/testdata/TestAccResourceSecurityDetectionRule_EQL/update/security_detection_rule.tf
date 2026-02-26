variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "eql"
  query       = "process where process.name == \"powershell.exe\""
  language    = "eql"
  enabled     = true
  description = "Updated test EQL security detection rule"
  severity    = "critical"
  risk_score  = 90
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["winlogbeat-*"]

  tiebreaker_field = "@timestamp"
  author           = ["Test Author"]
  tags             = ["test", "eql", "automation"]
  license          = "Elastic License v2"

  rule_name_override                   = "Updated Custom EQL Rule Name"
  timestamp_override                   = "process.end"
  timestamp_override_fallback_disabled = true

  filters = jsonencode([
    {
      "exists" = {
        "field" = "process.code_signature.trusted"
      }
    },
    {
      "term" = {
        "host.os.family" = "windows"
      }
    }
  ])

  investigation_fields = ["process.name", "process.executable", "process.parent.name"]

  risk_score_mapping = [
    {
      field      = "process.parent.name"
      operator   = "equals"
      value      = "cmd.exe"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "windows"
      version     = "2.0.0"
      integration = "system"
    }
  ]

  required_fields = [
    {
      name = "process.parent.name"
      type = "keyword"
    },
    {
      name = "event.category"
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
    group_by                = ["process.parent.name", "host.name"]
    duration                = "45m"
    missing_fields_strategy = "doNotSuppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "eql_response_pack"
        timeout = 450
        ecs_mapping = {
          "process.executable"  = "executable_path"
          "process.parent.name" = "parent_name"
        }
      }
    }
  ]
}

