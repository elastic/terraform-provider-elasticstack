variable "suffix" {
  description = "Random suffix to make the stream name unique."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "query" {
  name        = "logs.otel.testacc${var.suffix}.view"
  space_id    = "default"
  description = "Test query stream"

  query_config = {
    esql = "FROM logs* | LIMIT 10"
  }
}
