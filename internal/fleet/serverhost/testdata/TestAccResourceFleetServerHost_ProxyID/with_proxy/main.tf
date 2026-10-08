provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_proxy" "test" {
  name     = "Server Host Proxy ${var.suffix}"
  proxy_id = "server-host-proxy-${var.suffix}"
  url      = "https://proxy.example.com:3128"
}

resource "elasticstack_fleet_proxy" "other" {
  name     = "Server Host Other Proxy ${var.suffix}"
  proxy_id = "server-host-other-proxy-${var.suffix}"
  url      = "https://proxy2.example.com:3128"
}

resource "elasticstack_fleet_server_host" "test" {
  name    = "Proxy Server Host ${var.suffix}"
  host_id = "server-host-proxy-${var.suffix}"
  default = false
  hosts   = ["https://fleet-server:8220"]
  proxy_id = elasticstack_fleet_proxy.test.proxy_id
}
