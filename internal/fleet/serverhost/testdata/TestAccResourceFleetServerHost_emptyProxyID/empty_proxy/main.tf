provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_server_host" "test" {
  name     = "Empty Proxy Server Host"
  hosts    = ["https://fleet-server:8220"]
  proxy_id = ""
}
