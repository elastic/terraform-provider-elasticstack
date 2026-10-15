provider "elasticstack" {
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

resource "elasticstack_fleet_agent_download_source" "example" {
  name    = "Agent Download Source example"
  host    = "https://artifacts.elastic.co/downloads/elastic-agent"
  default = false

  # The ssl block requires Elastic Stack 9.1.0 or higher.
  ssl = {
    certificate_authorities = [trimspace(local.example_ca)]
    certificate             = trimspace(local.example_client_cert)
    key                     = trimspace(local.example_client_key)
  }
}
