variable "name" {
  type = string
}

variable "connector_name" {
  type = string
}

variable "connector_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_id      = var.connector_id
  connector_type_id = ".cases-webhook"

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
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  description = "Test state upgrade detection rule"
  type        = "query"
  severity    = "medium"
  risk_score  = 50
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  actions = [
    {
      action_type_id = ".cases-webhook"
      id             = elasticstack_kibana_action_connector.test.connector_id
      params = {
        message = "Test state upgrade alert"
      }
      group = "default"
      frequency = {
        notify_when = "onActiveAlert"
        summary     = true
        throttle    = "10m"
      }
    }
  ]
}
