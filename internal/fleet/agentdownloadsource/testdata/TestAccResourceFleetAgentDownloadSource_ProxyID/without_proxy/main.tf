provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_proxy" "test" {
  name     = "Agent Download Source Proxy ${var.suffix}"
  proxy_id = "agent-download-source-proxy-${var.suffix}"
  url      = "https://proxy.example.com:3128"
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Proxy Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-proxy-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-proxy"
  space_ids = ["default"]
}
