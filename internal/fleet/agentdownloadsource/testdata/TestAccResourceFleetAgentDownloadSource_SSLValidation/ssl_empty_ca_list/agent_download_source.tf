provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Agent Download Source"
  source_id = "agent-download-source"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]

  ssl = {
    certificate_authorities = []
  }
}
