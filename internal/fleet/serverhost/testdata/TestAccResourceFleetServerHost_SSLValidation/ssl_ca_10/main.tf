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
    certificate_authorities = ["PEM-CA-1", "PEM-CA-2", "PEM-CA-3", "PEM-CA-4", "PEM-CA-5", "PEM-CA-6", "PEM-CA-7", "PEM-CA-8", "PEM-CA-9", "PEM-CA-10"]
    client_auth             = "required"
  }
}
