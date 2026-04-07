provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Replace Source ID Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-replaced-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-replaced"
  space_ids = ["default"]
}
