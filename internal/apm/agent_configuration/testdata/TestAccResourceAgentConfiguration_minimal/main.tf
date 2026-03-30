variable "service_name" {
  description = "The APM service name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_agent_configuration" "test_config" {
  service_name = var.service_name
  settings = {
    "transaction_sample_rate" = "0.5"
  }
}
