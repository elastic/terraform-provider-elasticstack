variable "connector_name" {
  type = string
}

variable "space_id_a" {
  type = string
}

variable "space_id_b" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "a" {
  space_id = var.space_id_a
  name     = var.space_id_a
}

resource "elasticstack_kibana_space" "b" {
  space_id = var.space_id_b
  name     = var.space_id_b
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".index"
  space_id          = elasticstack_kibana_space.a.space_id
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
}
