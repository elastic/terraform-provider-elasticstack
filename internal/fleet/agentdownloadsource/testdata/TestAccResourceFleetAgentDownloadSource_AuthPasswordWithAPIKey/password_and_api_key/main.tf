provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name = "auth-password-and-api-key"
  host = "https://artifacts.example.com/elastic-agent"

  auth = {
    username = "svc"
    password = "secret"
    api_key  = "secret-key"
  }
}
