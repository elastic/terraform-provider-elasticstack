provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "client_auth" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name    = "SSL Server Host ${var.suffix}"
  host_id = "server-host-ssl-${var.suffix}"
  default = false
  hosts   = ["https://fleet-server:8220"]

  ssl = {
    client_auth = var.client_auth
  }
}
