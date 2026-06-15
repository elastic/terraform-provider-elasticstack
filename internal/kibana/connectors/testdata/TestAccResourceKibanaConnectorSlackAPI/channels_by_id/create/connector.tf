variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".slack_api"
  config = jsonencode({
    allowedChannels = [
      { id = "C0123456789" },
      { id = "C9876543210" },
    ]
  })
  secrets = jsonencode({
    token = "xoxb-test-token-for-slack-api"
  })
}
