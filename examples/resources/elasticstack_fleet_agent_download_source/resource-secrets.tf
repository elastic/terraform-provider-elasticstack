provider "elasticstack" {
  kibana {}
}

# Write-only attributes are never stored in Terraform state and require Terraform 1.11 or later.
# Bump the *_wo value (or change the secret) to have the provider detect the change and update Fleet.
resource "elasticstack_fleet_agent_download_source" "example" {
  name    = "Private Agent Download Source"
  host    = "https://artifacts.example.com/downloads/elastic-agent"
  default = false

  auth = {
    username    = "svc-account"
    password_wo = var.download_source_password
  }

  # The ssl block requires Elastic Stack 9.1.0 or higher.
  ssl = {
    certificate = trimspace(var.client_certificate)
    key_wo      = var.client_key
  }
}

variable "download_source_password" {
  type      = string
  sensitive = true
}

variable "client_certificate" {
  type = string
}

variable "client_key" {
  type      = string
  sensitive = true
}
