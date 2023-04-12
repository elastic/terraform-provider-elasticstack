provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_kibana_action_connector" "example" {
  name = "%s"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
  connector_type_id = ".index"
}
