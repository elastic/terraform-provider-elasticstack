variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = "agentbuilder-test-${var.tool_id}"
  deletion_protection = false
}

resource "elasticstack_kibana_agentbuilder_tool" "test_index_search" {
  depends_on = [elasticstack_elasticsearch_index.test_index]

  tool_id     = var.tool_id
  type        = "index_search"
  description = "Test index search tool"
  configuration = jsonencode({
    pattern = "agentbuilder-test-*"
  })
}
