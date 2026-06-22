variable "suffix" {
  description = "Random suffix to make the stream name unique."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "wired" {
  name        = "logs.otel.testacc${var.suffix}"
  space_id    = "default"
  description = "Wired stream with attached query"

  wired_config = {
    processing_steps = [
      jsonencode({
        action   = "grok"
        from     = "message"
        patterns = ["%%{GREEDYDATA:attributes.msg}"]
      })
    ]

    lifecycle_json     = jsonencode({ dsl = { data_retention = "30d" } })
    failure_store_json = jsonencode({ disabled = {} })

    index_number_of_shards   = 1
    index_number_of_replicas = 0
    index_refresh_interval   = "5s"
  }

  queries = [
    {
      id          = "testacc-query-${var.suffix}"
      title       = "Test Query"
      description = "A test query for acceptance testing"
      esql        = "FROM logs.otel.testacc${var.suffix} | LIMIT 10"
    }
  ]
}
