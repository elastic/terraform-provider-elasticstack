provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Empty Space IDs Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-empty-space-ids-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-empty-space-ids"
  space_ids = []
}
