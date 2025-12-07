variable "connector_name" {
  description = "The connector name"
  type        = string
}

variable "connector_id" {
  description = "Connector ID"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = "Updated ${var.connector_name}"
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
