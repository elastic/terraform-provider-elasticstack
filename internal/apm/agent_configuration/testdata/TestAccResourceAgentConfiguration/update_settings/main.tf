variable "service_name" {
  description = "The APM service name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_agent_configuration" "test_config" {
  service_name        = var.service_name
  service_environment = "production"
  agent_name          = "java"
  settings = {
    "transaction_sample_rate" = "0.8"
    "log_level"               = "debug"
  }
}
