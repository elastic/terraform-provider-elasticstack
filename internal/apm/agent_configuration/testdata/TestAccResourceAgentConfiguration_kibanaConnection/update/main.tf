variable "service_name" {
  description = "The APM service name"
  type        = string
}

variable "kibana_endpoints" {
  description = "Kibana base URLs for the entity-local connection block"
  type        = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Same as TestAccResourceAgentConfiguration/update, with entity-local kibana_connection
resource "elasticstack_apm_agent_configuration" "test_config" {
  service_name        = var.service_name
  service_environment = "production"
  agent_name          = "java"
  settings = {
    "transaction_sample_rate" = "0.8"
    "capture_body"            = "off"
  }

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.username != "" ? var.username : null
    password  = var.password != "" ? var.password : null
  }
}
