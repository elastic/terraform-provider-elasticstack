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
  name      = "Auth API Key Import Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-api-key-import-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-api-key-import"
  space_ids = ["default"]

  auth = {
    api_key = var.api_key
    headers = [{
      key   = "X-A"
      value = "1"
    }]
  }
}
