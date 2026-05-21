provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Placeholder PEM material for illustration only — replace with real certificates in production.
locals {
  example_ca          = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIBkTCB+wIJAKHHCgV4Jh0FMA0GCSqGGSIb3DQEBCwUAMBExCzAJBgNVBAYTAlVT
    -----END CERTIFICATE-----
  EOT
  example_client_cert = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIBkTCB+wIJAKHHCgV4Jh0FMA0GCSqGGSIb3DQEBCwUAMBExCzAJBgNVBAYTAlVT
    -----END CERTIFICATE-----
  EOT
  example_client_key  = <<-EOT
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEA0
    -----END RSA PRIVATE KEY-----
  EOT
}

resource "elasticstack_fleet_output" "logstash" {
  name                 = "Logstash Output"
  output_id            = "logstash-output"
  type                 = "logstash"
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "logstash.example.com:5044",
  ]

  ssl = {
    certificate_authorities = [trimspace(local.example_ca)]
    certificate             = trimspace(local.example_client_cert)
    key                     = trimspace(local.example_client_key)
  }
}
