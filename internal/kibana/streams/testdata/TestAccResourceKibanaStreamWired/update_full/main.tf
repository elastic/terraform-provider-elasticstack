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
  description = "Fully-configured wired stream"

  wired_config = {
    # Processing step — streamlang grok format
    processing_steps = [
      jsonencode({
        action   = "grok"
        from     = "message"
        patterns = ["%%{GREEDYDATA:attributes.msg}"]
      })
    ]

    # Lifecycle: retain data for 30 days
    lifecycle_json = jsonencode({
      dsl = { data_retention = "30d" }
    })

    # Failure store: disabled
    failure_store_json = jsonencode({ disabled = {} })

    # Index settings
    index_number_of_shards   = 1
    index_number_of_replicas = 0
    index_refresh_interval   = "5s"
  }

}
