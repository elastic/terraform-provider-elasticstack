provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Remove Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-remove-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-remove"
  space_ids = ["default"]
}
