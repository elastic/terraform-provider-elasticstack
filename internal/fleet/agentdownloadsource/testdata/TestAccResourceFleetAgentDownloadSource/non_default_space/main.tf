provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "non_default_space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.non_default_space_id
  name     = "Fleet Agent Download Source ${var.suffix}"
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Non Default Space Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-space-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-space"
  space_ids = [elasticstack_kibana_space.test.space_id]
}
