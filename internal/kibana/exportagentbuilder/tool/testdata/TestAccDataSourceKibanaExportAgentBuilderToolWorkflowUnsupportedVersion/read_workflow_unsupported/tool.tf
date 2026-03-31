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
  description = "ES|QL tool"
  configuration = jsonencode({
    query = "FROM logs-* | LIMIT ?limit"
    params = {
      limit = {
        type        = "integer"
        description = "Maximum number of results to return"
      }
    }
  })
}

data "elasticstack_kibana_agentbuilder_export_tool" "test" {
  id               = elasticstack_kibana_agentbuilder_tool.test.id
  include_workflow = true
}
