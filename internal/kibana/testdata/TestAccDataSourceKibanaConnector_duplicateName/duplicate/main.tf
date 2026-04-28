provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "first" {
  name              = var.connector_name
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://hooks.example.com/first"
  })
}

resource "elasticstack_kibana_action_connector" "second" {
  name              = var.connector_name
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://hooks.example.com/second"
  })
  depends_on = [elasticstack_kibana_action_connector.first]
}

data "elasticstack_kibana_action_connector" "test" {
  name       = var.connector_name
  depends_on = [elasticstack_kibana_action_connector.second]
}
