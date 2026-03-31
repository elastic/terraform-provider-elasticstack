variable "suffix" {
  description = "Random suffix to make the stream name unique."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "parent" {
  name     = "logs.otel.testacc-w${var.suffix}"
  space_id = "default"

  wired_config = {}
}

resource "elasticstack_kibana_stream" "query" {
  name        = "logs.otel.testacc-w${var.suffix}.view"
  space_id    = "default"
  description = "Updated query stream"

  query_config = {
    esql = "FROM $.logs.otel.testacc-w${var.suffix} | WHERE @timestamp > NOW() - 1 HOUR | LIMIT 10"
  }

  depends_on = [elasticstack_kibana_stream.parent]
}
