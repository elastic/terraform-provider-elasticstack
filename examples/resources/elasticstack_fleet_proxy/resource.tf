provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_proxy" "example" {
  name = "My Proxy"
  url  = "https://proxy.example.com:3128"

  certificate             = "-----BEGIN CERTIFICATE-----\n..."
  certificate_key         = "-----BEGIN PRIVATE KEY-----\n..."
  certificate_authorities = "-----BEGIN CERTIFICATE-----\n..."

  proxy_headers = {
    "X-Custom-Header" = "my-value"
  }
}
