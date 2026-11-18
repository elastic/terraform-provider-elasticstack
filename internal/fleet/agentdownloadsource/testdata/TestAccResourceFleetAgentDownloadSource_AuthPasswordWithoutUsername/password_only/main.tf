provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name = "auth-password-only"
  host = "https://artifacts.example.com/elastic-agent"

  auth = {
    password = "secret"
  }
}
