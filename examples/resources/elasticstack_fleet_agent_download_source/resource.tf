provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "example" {
  name    = "Agent Download Source example"
  host    = "https://artifacts.elastic.co/downloads/elastic-agent"
  default = false
}

