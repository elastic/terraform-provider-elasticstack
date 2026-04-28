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
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test_esql" {
  tool_id     = var.tool_id
  type        = "esql"
  description = "Updated ES|QL tool (kibana_connection)"
  tags        = ["test", "kibana_conn", "updated"]
  configuration = jsonencode({
    query = "FROM logs-* | WHERE @timestamp >= ?startTime | LIMIT ?limit"
    params = {
      limit = {
        type        = "integer"
        description = "Maximum number of results to return"
      }
      startTime = {
        type        = "date"
        description = "Start time in ISO format"
      }
    }
  })

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.username != "" ? var.username : null
    password  = var.password != "" ? var.password : null
  }
}
