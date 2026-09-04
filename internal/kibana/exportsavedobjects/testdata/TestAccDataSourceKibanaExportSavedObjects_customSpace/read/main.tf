variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = "test-export-connector-custom-space"
  connector_type_id = ".index"
  space_id          = elasticstack_kibana_space.test.space_id
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    }
  ]
}
