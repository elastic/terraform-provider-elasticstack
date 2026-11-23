provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_server_host" "test" {
  name    = "SSL Server Host ${var.suffix}"
  host_id = "server-host-ssl-${var.suffix}"
  default = false
  hosts   = ["https://fleet-server:8220"]

  ssl = {
    certificate             = "PEM-CERT"
    key                     = "PEM-KEY"
    certificate_authorities = ["PEM-CA-1", "PEM-CA-2"]
    client_auth             = "sometimes"
  }
}
