variable "connector_name" {
  type = string
}

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
  name              = var.connector_name
  connector_type_id = ".index"
  space_id          = elasticstack_kibana_space.test.space_id
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
}
