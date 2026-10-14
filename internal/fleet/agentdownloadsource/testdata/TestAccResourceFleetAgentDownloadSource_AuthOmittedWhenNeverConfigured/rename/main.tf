provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Omitted Renamed Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-omitted-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-omitted"
  space_ids = ["default"]
}
