provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Auth Echo Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-echo-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-echo"
  space_ids = ["default"]

  auth = {
    username = "svc-rotated"
    password = var.password
    headers = [{
      key   = "X-B"
      value = "2"
    }]
  }
}
