provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = "test-export-connector-multi"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://example.com"
  })
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id = "default"
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    },
    {
      type = "space",
      id   = "default"
    }
  ]
}
