provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "first" {
  name    = "Data Source Duplicate Server Host ${var.suffix}"
  host_id = "ds-duplicate-a-${var.suffix}"
  hosts   = ["https://fleet-server-1:8220"]
}

resource "elasticstack_fleet_server_host" "second" {
  name    = "Data Source Duplicate Server Host ${var.suffix}"
  host_id = "ds-duplicate-b-${var.suffix}"
  hosts   = ["https://fleet-server-2:8220"]
}

data "elasticstack_fleet_server_host" "test" {
  name = "Data Source Duplicate Server Host ${var.suffix}"

  depends_on = [
    elasticstack_fleet_server_host.first,
    elasticstack_fleet_server_host.second,
  ]
}
