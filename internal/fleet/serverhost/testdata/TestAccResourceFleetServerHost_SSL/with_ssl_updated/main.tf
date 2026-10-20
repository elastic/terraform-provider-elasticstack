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
    certificate             = "PEM-CERT-2"
    key                     = "PEM-KEY-2"
    certificate_authorities = ["PEM-CA-3", "PEM-CA-4", "PEM-CA-5"]
    client_auth             = "optional"
  }
}
