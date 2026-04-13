variable "tool_id" {
  description = "The tool ID"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Tool Export"
  description = "Space for testing agent builder tool exports"
}

resource "elasticstack_kibana_agentbuilder_tool" "test" {
  tool_id  = var.tool_id
  space_id = elasticstack_kibana_space.test.space_id
  type     = "esql"
  configuration = jsonencode({
    query  = "FROM logs | LIMIT 10"
    params = {}
  })
}

data "elasticstack_kibana_agentbuilder_tool" "test" {
  id = elasticstack_kibana_agentbuilder_tool.test.id
}
