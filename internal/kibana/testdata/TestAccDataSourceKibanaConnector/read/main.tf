provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "slack" {
  name              = "myconnector"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://internet.com"
  })
}

data "elasticstack_kibana_action_connector" "myconnector" {
  name = elasticstack_kibana_action_connector.slack.name
}
