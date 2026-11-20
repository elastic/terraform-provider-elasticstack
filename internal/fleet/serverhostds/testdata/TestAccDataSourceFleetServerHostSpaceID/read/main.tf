provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Data Source Server Host Space ${var.suffix}"
}

resource "elasticstack_fleet_server_host" "test" {
  name      = "Data Source Space Server Host ${var.suffix}"
  hosts     = ["https://fleet-server:8220"]
  space_ids = [elasticstack_kibana_space.test.space_id]
}

data "elasticstack_fleet_server_host" "omitted" {
  name = elasticstack_fleet_server_host.test.name
}

data "elasticstack_fleet_server_host" "same_space" {
  name     = elasticstack_fleet_server_host.test.name
  space_id = elasticstack_kibana_space.test.space_id
}

data "elasticstack_fleet_server_host" "default_space" {
  host_id  = elasticstack_fleet_server_host.test.host_id
  space_id = "default"
}
