provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name  = "Data Source SSL Server Host ${var.suffix}"
  hosts = ["https://fleet-server:8220"]

  ssl = {
    certificate = "PEM-CERT"
    key         = "PEM-KEY"
  }
}

data "elasticstack_fleet_server_host" "test" {
  name = elasticstack_fleet_server_host.test.name
}
