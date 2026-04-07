provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  proxy_id  = "proxy-123"
  space_ids = ["default"]
}
