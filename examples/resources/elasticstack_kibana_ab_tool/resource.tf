provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_ab_tool" "my_esql_tool" {
  id          = "my-esql-tool"
  type        = "esql"
  description = "Search logs using ES|QL"
  tags        = ["logs", "esql"]

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
