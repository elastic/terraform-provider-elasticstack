provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_proxy" "example" {
  name = "Corporate Proxy"
  url  = "https://proxy.example.com:8080"
}
