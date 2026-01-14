resource "elasticstack_fleet_proxy" "secure_proxy" {
  name = "Secure Corporate Proxy"
  url  = "https://proxy.example.com:8443"

  certificate             = file("${path.module}/certs/client-cert.pem")
  certificate_authorities = file("${path.module}/certs/ca.pem")
  certificate_key         = file("${path.module}/certs/client-key.pem")
}
