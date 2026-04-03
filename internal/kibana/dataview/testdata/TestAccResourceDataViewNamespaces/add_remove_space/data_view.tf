variable "index_name" { type = string }
variable "space1"     { type = string }
variable "space2"     { type = string }
variable "space3"     { type = string }

resource "elasticstack_kibana_space" "space1" {
  space_id = var.space1
  name     = var.space1
}

resource "elasticstack_kibana_space" "space2" {
  space_id = var.space2
  name     = var.space2
}

resource "elasticstack_kibana_space" "space3" {
  space_id = var.space3
  name     = var.space3
}

resource "elasticstack_kibana_data_view" "ns_dv" {
  data_view = {
    title      = var.index_name
    namespaces = ["default", var.space1, var.space3]
  }
  depends_on = [
    elasticstack_kibana_space.space1,
    elasticstack_kibana_space.space2,
    elasticstack_kibana_space.space3,
  ]
}