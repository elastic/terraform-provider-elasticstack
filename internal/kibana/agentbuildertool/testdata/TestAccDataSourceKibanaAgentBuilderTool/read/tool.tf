variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test" {
  tool_id     = var.tool_id
  type        = "esql"
  description = "Test ESQL tool"
  tags        = ["test"]
  configuration = jsonencode({
    query  = "FROM logs | LIMIT 10"
    params = {}
  })
}

data "elasticstack_kibana_agentbuilder_tool" "test" {
  id = elasticstack_kibana_agentbuilder_tool.test.id
}
