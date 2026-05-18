variable "connector_name" {
  description = "The connector name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".index"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
  secrets = jsonencode({})
}

data "elasticstack_kibana_action_connector" "test" {
  name = elasticstack_kibana_action_connector.test.name
}
