variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threat_match"
  query       = "destination.ip:* OR source.ip:*"
  language    = "kuery"
  enabled     = true
  description = "Updated test threat match security detection rule"
  severity    = "critical"
  risk_score  = 95
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*", "network-*"]

  namespace    = "updated-threat-match-namespace"
  threat_index = ["threat-intel-*", "ioc-*"]
  threat_query = "threat.indicator.type:(ip OR domain)"

  author  = ["Test Author"]
  tags    = ["test", "threat-match", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom Threat Match Rule Name"
  timestamp_override                   = "threat.indicator.last_seen"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "regexp" = {
        "destination.domain" = ".*\\.suspicious\\.com"
      }
    }
  ])

  investigation_fields = ["destination.ip", "source.ip", "threat.indicator.type"]

  threat_mapping = [
    {
      entries = [
        {
          field = "destination.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    },
    {
      entries = [
        {
          field = "source.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    }
  ]

  risk_score_mapping = [
    {
      field      = "threat.indicator.confidence"
      operator   = "equals"
      value      = "high"
      risk_score = 100
    }
  ]

  related_integrations = [
    {
      package     = "threat_intel"
      version     = "2.0.0"
      integration = "indicators"
    }
  ]

  required_fields = [
    {
      name = "destination.ip"
      type = "ip"
    },
    {
      name = "source.ip"
      type = "ip"
    },
    {
      name = "threat.indicator.ip"
      type = "ip"
    }
  ]

  severity_mapping = [
    {
      field    = "threat.indicator.confidence"
      operator = "equals"
      value    = "high"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "threat_intel_investigation"
        timeout        = 450
        ecs_mapping = {
          "source.ip"             = "src_ip"
          "destination.ip"        = "dest_ip"
          "threat.indicator.type" = "threat_type"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "kill-process"
        comment = "Kill processes communicating with known threat indicators"
        config = {
          field     = "process.entity_id"
          overwrite = false
        }
      }
    }
  ]
}

