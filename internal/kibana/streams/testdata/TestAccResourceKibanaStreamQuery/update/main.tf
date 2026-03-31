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
  description = "Updated query stream"

  query_config = {
    esql = "FROM logs* | WHERE @timestamp > NOW() - 1 HOUR | LIMIT 10"
  }
}
