variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                 = var.name
  type                 = "new_terms"
  query                = "user.name:* AND source.ip:*"
  language             = "kuery"
  enabled              = true
  description          = "Updated test new terms security detection rule"
  severity             = "high"
  risk_score           = 75
  from                 = "now-6m"
  to                   = "now"
  interval             = "5m"
  index                = ["logs-*", "audit-*"]
  new_terms_fields     = ["user.name", "source.ip"]
  history_window_start = "now-30d"

  author  = ["Test Author"]
  tags    = ["test", "new-terms", "automation"]
  license = "Elastic License v2"

  rule_name_override                   = "Updated Custom New Terms Rule Name"
  timestamp_override                   = "user.last_login"
  timestamp_override_fallback_disabled = false

  filters = jsonencode([
    {
      "geo_distance" = {
        "distance" = "1000km"
        "source.geo.location" = {
          "lat" = 40.12
          "lon" = -71.34
        }
      }
    }
  ])

  investigation_fields = ["user.name", "user.type", "source.ip", "user.roles"]

  risk_score_mapping = [
    {
      field      = "user.roles"
      operator   = "equals"
      value      = "admin"
      risk_score = 95
    },
    {
      field      = "source.geo.country_name"
      operator   = "equals"
      value      = "CN"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "security"
      version     = "2.0.0"
      integration = "users"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "source.ip"
      type = "ip"
    },
    {
      name = "user.roles"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.roles"
      operator = "equals"
      value    = "admin"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "admin_user_investigation"
        timeout        = 800
        ecs_mapping = {
          "user.roles" = "roles"
          "source.ip"  = "source_ip"
          "user.name"  = "username"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to new admin user activity from suspicious IP"
      }
    }
  ]
}

