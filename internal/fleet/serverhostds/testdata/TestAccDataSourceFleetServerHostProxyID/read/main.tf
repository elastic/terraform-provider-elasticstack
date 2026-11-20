provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_proxy" "test" {
  name     = "Data Source Server Host Proxy ${var.suffix}"
  proxy_id = "ds-server-host-proxy-${var.suffix}"
  url      = "https://proxy.example.com:3128"
}

resource "elasticstack_fleet_server_host" "test" {
  name     = "Data Source Proxy Server Host ${var.suffix}"
  hosts    = ["https://fleet-server:8220"]
  proxy_id = elasticstack_fleet_proxy.test.proxy_id
}

data "elasticstack_fleet_server_host" "test" {
  host_id = elasticstack_fleet_server_host.test.host_id
}
