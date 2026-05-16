variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test_index_search" {
  tool_id     = var.tool_id
  type        = "index_search"
  description = "Test index search tool"
  configuration = jsonencode({
    pattern = "logs-test-*"
  })
}
