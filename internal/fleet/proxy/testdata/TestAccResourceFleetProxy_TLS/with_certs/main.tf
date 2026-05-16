variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_proxy" "test_proxy" {
  name = var.name
  url  = "https://proxy-tls.example.com:3128"

  certificate             = "PEM-CERT"
  certificate_authorities = "PEM-CA"
  certificate_key         = "PEM-KEY"
}
