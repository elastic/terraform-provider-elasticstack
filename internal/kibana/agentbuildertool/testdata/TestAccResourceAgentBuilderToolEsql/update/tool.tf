variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test_esql" {
  tool_id     = var.tool_id
  type        = "esql"
  description = "Updated ES|QL tool"
  tags        = ["test", "esql", "updated"]
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
}
