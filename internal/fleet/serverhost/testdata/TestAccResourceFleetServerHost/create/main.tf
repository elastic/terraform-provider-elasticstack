variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = var.name
  host_id = "fleet-server-host-id"
  default = false
  hosts = [
    "https://fleet-server:8220"
  ]
}
