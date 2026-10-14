provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "example" {
  name    = "Private Agent Download Source"
  host    = "https://artifacts.example.com/downloads/elastic-agent"
  default = false

  auth = {
    username = "svc-account"
    password = "example-password"
    headers = [
      {
        key   = "X-Custom-Header"
        value = "value"
      },
    ]
  }
}
