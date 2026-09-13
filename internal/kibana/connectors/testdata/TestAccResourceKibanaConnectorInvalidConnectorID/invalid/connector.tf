resource "elasticstack_kibana_action_connector" "test" {
  name              = "invalid-connector-id-test"
  connector_type_id = ".index"
  connector_id      = "not-a-uuid"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
}
