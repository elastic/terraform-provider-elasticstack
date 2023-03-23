provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Test Output"
  type                 = "elasticsearch"
  config_yaml          = yamlencode({
    "ssl.verification_mode": "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts                = [
    "https://elasticsearch:9200"
  ]
}
