provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Headers Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-headers-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-headers"
  space_ids = ["default"]

  auth = {
    headers = [{
      key   = "X-A"
      value = "1"
    }]
  }
}
