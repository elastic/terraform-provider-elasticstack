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
  space_ids = ["default"]

  ssl = {
    certificate             = "PEM-CERT"
    key                     = "PEM-KEY"
    certificate_authorities = ["PEM-CA-1", "PEM-CA-2", "PEM-CA-3", "PEM-CA-4", "PEM-CA-5", "PEM-CA-6", "PEM-CA-7", "PEM-CA-8", "PEM-CA-9", "PEM-CA-10", "PEM-CA-11"]
  }
}
