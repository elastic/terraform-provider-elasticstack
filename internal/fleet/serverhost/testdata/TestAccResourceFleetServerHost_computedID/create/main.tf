variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_computed_id" {
  name  = var.name
  hosts = ["https://fleet-server:8220"]
}
