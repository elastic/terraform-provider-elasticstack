provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Rotate Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-rotate-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-rotate"
  space_ids = ["default"]

  auth = {
    username = "svc-account"
    password = var.password
  }
}
