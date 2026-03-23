variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = var.connector_name
  config = jsonencode({
    createIncidentJson                  = "{}"
    createIncidentResponseKey           = "key"
    createIncidentUrl                   = "https://www.elastic.co/"
    getIncidentResponseExternalTitleKey = "title"
    getIncidentUrl                      = "https://www.elastic.co/"
    updateIncidentJson                  = "{}"
    updateIncidentUrl                   = "https://www.elastic.co/"
    viewIncidentUrl                     = "https://www.elastic.co/"
  })
  secrets = jsonencode({
    user     = "user1"
    password = "password1"
  })
  connector_type_id = ".cases-webhook"
}
