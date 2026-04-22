provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Updated Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-${var.suffix}"
  default   = true
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-updated"
  space_ids = ["default"]
}
