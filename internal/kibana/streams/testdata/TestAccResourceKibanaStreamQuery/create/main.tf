variable "suffix" {
  description = "Random suffix to make the stream name unique."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "query" {
  name        = "logs.otel.testacc-q${var.suffix}"
  space_id    = "default"
  description = "Test query stream"

  query_config = {
    esql = "FROM $.logs.otel | LIMIT 10"
  }
}
