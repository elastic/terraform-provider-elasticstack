variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threat_match"
  query       = "destination.ip:*"
  language    = "kuery"
  enabled     = true
  description = "Test threat match security detection rule"
  severity    = "high"
  risk_score  = 80
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  namespace                            = "threat-match-namespace"
  rule_name_override                   = "Custom Threat Match Rule Name"
  timestamp_override                   = "threat.indicator.first_seen"
  timestamp_override_fallback_disabled = true

  threat_index = ["threat-intel-*"]
  threat_query = "threat.indicator.type:ip"

  filters = jsonencode([
    {
      "bool" = {
        "must_not" = [
          {
            "term" = {
              "destination.ip" = "127.0.0.1"
            }
          }
        ]
      }
    }
  ])

  investigation_fields = ["destination.ip", "source.ip"]

  threat_mapping = [
    {
      entries = [
        {
          field = "destination.ip"
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
      value      = "medium"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "threat_intel"
      version     = "1.0.0"
      integration = "indicators"
    }
  ]

  required_fields = [
    {
      name = "destination.ip"
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
      severity = "high"
    }
  ]

  alert_suppression = {
    group_by                = ["destination.ip", "source.ip"]
    duration                = "1h"
    missing_fields_strategy = "doNotSuppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM listening_ports WHERE address = '{{destination.ip}}';"
        timeout = 300
        ecs_mapping = {
          "destination.ip"              = "dest_ip"
          "threat.indicator.ip"         = "threat_ip"
          "threat.indicator.confidence" = "confidence"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to threat match on destination IP"
      }
    }
  ]
}

