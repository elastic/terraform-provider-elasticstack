provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Default Toggle Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-toggle-${var.suffix}"
  default   = true
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}
