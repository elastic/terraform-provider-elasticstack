variable "agent_id" {
  description = "The agent ID"
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

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Test Agent (kibana_connection)"
  description  = "An agent exported through an entity-local Kibana connection"
  instructions = "Use the scoped Kibana connection."
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id = elasticstack_kibana_agentbuilder_agent.test.agent_id

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
