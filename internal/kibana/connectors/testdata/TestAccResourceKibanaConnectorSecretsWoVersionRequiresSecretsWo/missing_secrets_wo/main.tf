resource "elasticstack_kibana_action_connector" "test" {
  name              = "also-requires-test"
  connector_type_id = ".pagerduty"
  config = jsonencode({
    apiUrl = "https://events.pagerduty.com/v2/enqueue"
  })
  secrets_wo_version = "v1"
}
