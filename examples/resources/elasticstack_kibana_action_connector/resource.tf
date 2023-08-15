provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "example" {
  name = "%s"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
  connector_type_id = ".index"
}

resource "elasticstack_kibana_action_connector" "pagerduty-connector" {
  name              = "pagerduty"
  connector_type_id = ".pagerduty"
  config = jsonencode({
    apiUrl = "https://events.pagerduty.com/v2/enqueue"
  })
  secrets = jsonencode({
    routingKey = pagerduty_service_integration.kibana.integration_key
  })
}

resource "elasticstack_kibana_action_connector" "slack-connector" {
  name              = "slack"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "<your-webhookUrl>"
  })
}
