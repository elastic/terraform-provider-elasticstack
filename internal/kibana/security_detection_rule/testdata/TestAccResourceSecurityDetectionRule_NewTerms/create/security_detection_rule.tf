variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                 = var.name
  type                 = "new_terms"
  query                = "user.name:*"
  language             = "kuery"
  enabled              = true
  description          = "Test new terms security detection rule"
  severity             = "medium"
  risk_score           = 50
  from                 = "now-6m"
  to                   = "now"
  interval             = "5m"
  index                = ["logs-*"]
  new_terms_fields     = ["user.name"]
  history_window_start = "now-14d"

  namespace                            = "new-terms-namespace"
  rule_name_override                   = "Custom New Terms Rule Name"
  timestamp_override                   = "user.created"
  timestamp_override_fallback_disabled = true

  filters = jsonencode([
    {
      "bool" = {
        "should" = [
          {
            "wildcard" = {
              "user.domain" = "*.internal"
            }
          },
          {
            "term" = {
              "user.type" = "service_account"
            }
          }
        ]
      }
    }
  ])

  investigation_fields = ["user.name", "user.type"]

  risk_score_mapping = [
    {
      field      = "user.type"
      operator   = "equals"
      value      = "service_account"
      risk_score = 65
    }
  ]

  related_integrations = [
    {
      package     = "security"
      version     = "1.0.0"
      integration = "users"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "user.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.type"
      operator = "equals"
      value    = "service_account"
      severity = "medium"
    }
  ]

  alert_suppression = {
    group_by                = ["user.name", "user.type"]
    duration                = "20m"
    missing_fields_strategy = "doNotSuppress"
  }

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM last WHERE username = '{{user.name}}';"
        timeout = 350
        ecs_mapping = {
          "user.name" = "username"
          "user.type" = "user_type"
          "host.name" = "hostname"
        }
      }
    }
  ]
}

