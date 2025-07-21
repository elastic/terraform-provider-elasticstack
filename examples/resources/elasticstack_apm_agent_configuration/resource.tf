provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_apm_agent_configuration" "test_config" {
  service_name        = "my-service"
  service_environment = "production"
  agent_name          = "go"
  settings = {
    "transaction_sample_rate" = "0.5"
    "capture_body"            = "all"
  }
}
