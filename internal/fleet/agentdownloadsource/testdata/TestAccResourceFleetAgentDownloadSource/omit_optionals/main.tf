provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name = "No Optionals Agent Download Source ${var.suffix}"
  host = "https://artifacts.elastic.co/downloads/elastic-agent-no-optionals"
}
