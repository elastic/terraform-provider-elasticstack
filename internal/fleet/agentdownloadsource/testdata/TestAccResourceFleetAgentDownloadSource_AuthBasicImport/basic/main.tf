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
  name      = "Auth Basic Import Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-auth-basic-import-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-auth-basic-import"
  space_ids = ["default"]

  auth = {
    username = "svc-account"
    password = var.password
    headers = [{
      key   = "X-A"
      value = "1"
    }]
  }
}
