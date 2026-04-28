variable "exclude_export_details" {
  type = bool
}

variable "include_references_deep" {
  type = bool
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = "test-export-connector-boolopts"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://example.com"
  })
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id                = "default"
  exclude_export_details  = var.exclude_export_details
  include_references_deep = var.include_references_deep
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    }
  ]
}
