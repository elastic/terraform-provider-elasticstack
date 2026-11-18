provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name  = "Data Source Server Host ${var.suffix}"
  hosts = ["https://fleet-server-3:8220"]
}

data "elasticstack_fleet_server_host" "test" {
  name = elasticstack_fleet_server_host.test.name
}
