provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_proxy" "example" {
  name = "Example Proxy"
  url  = "https://proxy.example.com:3128"
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = "Test Host"
  default = false
  hosts = [
    "https://fleet-server:8220"
  ]
  proxy_id = elasticstack_fleet_proxy.example.proxy_id

  # Requires Elastic Stack 9.1.0 or later.
  ssl = {
    certificate             = "-----BEGIN CERTIFICATE-----\n..."
    key                     = "-----BEGIN PRIVATE KEY-----\n..."
    certificate_authorities = ["-----BEGIN CERTIFICATE-----\n..."]
    client_auth             = "required"
  }
}
