provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "slack" {
  name              = "myslackconnector"
  connector_type_id = ".slack"
  config            = "{}"
  secrets           = "{}"
}

data "elasticstack_kibana_action_connector" "example" {
  name              = elasticstack_kibana_action_connector.slack.name
  space_id          = "default"
  connector_type_id = ".slack"
}

output "connector_id" {
  value = data.elasticstack_kibana_action_connector.example.connector_id
}
