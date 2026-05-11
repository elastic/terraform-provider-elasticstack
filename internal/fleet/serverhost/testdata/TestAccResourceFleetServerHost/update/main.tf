variable "name" {
  type = string
}

variable "host_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = var.name
  host_id = var.host_id
  default = true
  hosts = [
    "https://fleet-server:8220",
    "https://fleet-server-2:8220"
  ]
}
