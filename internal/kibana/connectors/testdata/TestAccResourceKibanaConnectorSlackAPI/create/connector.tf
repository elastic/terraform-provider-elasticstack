variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".slack_api"
  config = jsonencode({
    allowedChannels = [
      { name = "#kar_testing" },
      { name = "#test-prod-alerts" },
    ]
  })
  secrets = jsonencode({
    token = "xoxb-test-token-for-slack-api"
  })
}
