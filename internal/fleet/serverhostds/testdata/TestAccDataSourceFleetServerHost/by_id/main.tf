provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name  = "Data Source Server Host ${var.suffix}"
  hosts = ["https://fleet-server-1:8220", "https://fleet-server-2:8220"]
}

data "elasticstack_fleet_server_host" "test" {
  host_id = elasticstack_fleet_server_host.test.host_id
}
