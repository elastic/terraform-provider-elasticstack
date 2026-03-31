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
  name        = "Test Space for Tools"
  description = "Space for testing agent builder tools"
}

resource "elasticstack_kibana_agentbuilder_tool" "test_esql" {
  tool_id  = var.tool_id
  space_id = elasticstack_kibana_space.test.space_id
  type     = "esql"
  configuration = jsonencode({
    query  = "FROM logs-* | LIMIT 10"
    params = {}
  })
}
