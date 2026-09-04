provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "suffix" {
  type = string
}

variable "second_space_id" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Space Update Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-space-update-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-space-update"
  space_ids = ["default"]
}
