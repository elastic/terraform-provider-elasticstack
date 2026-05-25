resource "elasticstack_kibana_action_connector" "test" {
  name              = "conflict-test"
  connector_type_id = ".pagerduty"
  config = jsonencode({
    apiUrl = "https://events.pagerduty.com/v2/enqueue"
  })
  secrets = jsonencode({
    routingKey = "plain-key"
  })
  secrets_wo = jsonencode({
    routingKey = "wo-key"
  })
}
