provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = "test-export-connector"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://example.com"
  })
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id                = "default"
  exclude_export_details  = true
  include_references_deep = true
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    }
  ]
}
