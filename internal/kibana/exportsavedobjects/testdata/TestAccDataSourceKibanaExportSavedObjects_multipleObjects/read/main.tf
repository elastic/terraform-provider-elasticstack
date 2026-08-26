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

resource "elasticstack_kibana_action_connector" "test2" {
  name              = "test-export-connector-multi-2"
  connector_type_id = ".server-log"
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id = "default"
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    },
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test2.connector_id
    }
  ]
}
