variable "policy_name" {
  description = "The integration policy name"
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

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "${var.policy_name}-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  integration_version = "8.14.0"
  preset              = "EDRComplete"

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }

  policy = {
    windows = {
      events = {
        process = true
        network = true
        file    = true
      }
      malware = {
        mode = "prevent"
      }
    }
    mac = {
      events = {
        process = true
      }
    }
    linux = {
      events = {
        process = true
        network = true
      }
    }
  }
}
