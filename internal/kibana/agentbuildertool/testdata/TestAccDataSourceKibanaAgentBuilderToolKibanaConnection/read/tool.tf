variable "tool_id" {
  description = "The tool ID"
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
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test" {
  tool_id     = var.tool_id
  type        = "esql"
  description = "Test ESQL tool (DS kibana_connection)"
  tags        = ["test", "ds-conn"]
  configuration = jsonencode({
    query  = "FROM logs | LIMIT 10"
    params = {}
  })
}

data "elasticstack_kibana_agentbuilder_tool" "test" {
  id = elasticstack_kibana_agentbuilder_tool.test.id

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
