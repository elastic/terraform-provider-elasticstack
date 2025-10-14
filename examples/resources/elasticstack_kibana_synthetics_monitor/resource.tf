provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_monitor" "my_monitor" {
  name      = "Example http monitor"
  space_id  = "default"
  schedule  = 10
  locations = ["us_west"]
  enabled   = false
  tags      = ["tag"]
  labels = {
    environment = "production"
    team        = "platform"
    service     = "web-app"
  }
  alert = {
    status = {
      enabled = true
    }
    tls = {
      enabled = false
    }
  }
  service_name = "example apm service"
  timeout      = 30
  http = {
    url                     = "http://localhost:8080"
    ssl_verification_mode   = "full"
    ssl_supported_protocols = ["TLSv1.2"]
    max_redirects           = "10"
    mode                    = "all"
    ipv4                    = true
    ipv6                    = true
  }
}
