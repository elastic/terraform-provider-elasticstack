resource "elasticstack_fleet_proxy" "fleet_proxy" {
  name = "Fleet Server Proxy"
  url  = "https://proxy.example.com:8080"
}

resource "elasticstack_fleet_server_host" "example" {
  name     = "Fleet Server"
  hosts    = ["https://fleet-server:8220"]
  proxy_id = elasticstack_fleet_proxy.fleet_proxy.proxy_id
}
