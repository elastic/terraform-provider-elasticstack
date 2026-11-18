provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name    = "Data Source Default Server Host ${var.suffix}"
  hosts   = ["https://fleet-server:8220"]
  default = true
}

data "elasticstack_fleet_server_host" "test" {
  default = true

  depends_on = [elasticstack_fleet_server_host.test]
}

data "elasticstack_fleet_server_host" "by_id" {
  host_id = elasticstack_fleet_server_host.test.host_id
}
