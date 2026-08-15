variable "connector_name" {
  description = "The connector name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "slack" {
  name              = var.connector_name
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://hooks.example.com/slack"
  })
}

resource "elasticstack_kibana_action_connector" "index" {
  name              = var.connector_name
  connector_type_id = ".index"
  config = jsonencode({
    index = ".kibana"
  })
  depends_on = [elasticstack_kibana_action_connector.slack]
}

data "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".index"
  depends_on        = [elasticstack_kibana_action_connector.index]
}
