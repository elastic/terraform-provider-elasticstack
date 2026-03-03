variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = "test connector 1"
  connector_id = "1d30b67b-f90b-4e28-87c2-137cba361509"

  config = jsonencode({
    createIncidentJson                  = "{}"
    createIncidentResponseKey           = "key"
    createIncidentUrl                   = "https://www.elastic.co/"
    getIncidentResponseExternalTitleKey = "title"
    getIncidentUrl                      = "https://www.elastic.co/"
    updateIncidentJson                  = "{}"
    updateIncidentUrl                   = "https://elasticsearch.com/"
    viewIncidentUrl                     = "https://www.elastic.co/"
    createIncidentMethod                = "put"
  })

  secrets = jsonencode({
    user     = "user2"
    password = "password2"
  })

  connector_type_id = ".cases-webhook"
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  description = "Updated test security detection rule with connector action"
  type        = "query"
  severity    = "high"
  risk_score  = 75
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  namespace   = "updated-connector-action-namespace"

  tags = ["test", "terraform"]

  risk_score_mapping = [
    {
      field      = "user.privileged"
      operator   = "equals"
      value      = "true"
      risk_score = 95
    }
  ]

  actions = [
    {
      action_type_id = ".cases-webhook"
      id             = elasticstack_kibana_action_connector.test.connector_id
      params = {
        message = "UPDATED CRITICAL Alert: Security event detected"
      }
      group = "default"
      frequency = {
        notify_when = "onActiveAlert"
        summary     = true
        throttle    = "5m"
      }
    }
  ]

  exceptions_list = [
    {
      id             = "test-action-exception"
      list_id        = "action-rule-exceptions"
      namespace_type = "single"
      type           = "detection"
    }
  ]
}

