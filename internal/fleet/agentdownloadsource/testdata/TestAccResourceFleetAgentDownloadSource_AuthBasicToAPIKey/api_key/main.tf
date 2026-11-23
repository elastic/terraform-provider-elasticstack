provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "api_key" {
  type      = string
  sensitive = true
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Replace Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-replace-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-replace"
  space_ids = ["default"]

  auth = {
    api_key = var.api_key
  }
}
