provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name = "auth-headers-over-max"
  host = "https://artifacts.example.com/elastic-agent"

  auth = {
    headers = [for i in range(101) : {
      key   = "X-${i}"
      value = tostring(i)
    }]
  }
}
