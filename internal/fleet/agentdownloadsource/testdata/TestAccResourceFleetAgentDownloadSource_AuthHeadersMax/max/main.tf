provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Headers Max Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-headers-max-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-headers-max"
  space_ids = ["default"]

  auth = {
    headers = [for i in range(100) : {
      key   = "X-${i}"
      value = tostring(i)
    }]
  }
}
