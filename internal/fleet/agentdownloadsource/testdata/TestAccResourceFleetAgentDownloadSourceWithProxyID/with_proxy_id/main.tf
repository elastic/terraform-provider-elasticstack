provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "proxy_id" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Proxy ID Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-proxy-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
  proxy_id  = var.proxy_id
}
