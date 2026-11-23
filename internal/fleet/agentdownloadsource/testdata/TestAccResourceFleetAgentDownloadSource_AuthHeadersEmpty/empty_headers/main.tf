provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Empty Headers Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-empty-headers-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-empty-headers"
  space_ids = ["default"]

  auth = {
    headers = []
  }
}
