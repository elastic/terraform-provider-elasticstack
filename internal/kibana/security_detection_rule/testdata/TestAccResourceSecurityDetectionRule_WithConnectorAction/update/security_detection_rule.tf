variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

variable "connector_name" {
  type = string
}

variable "connector_id" {
  type = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = var.connector_name
  connector_id = var.connector_id

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

