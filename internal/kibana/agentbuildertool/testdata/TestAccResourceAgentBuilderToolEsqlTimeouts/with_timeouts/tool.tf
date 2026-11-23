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
  description = "Test ES|QL tool (timeouts)"
  configuration = jsonencode({
    query  = "FROM logs-* | LIMIT 10"
    params = {}
  })

  timeouts = {
    create = "5m"
  }
}
