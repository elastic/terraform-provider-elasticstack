# Classic streams adopt pre-existing Elasticsearch data streams.
# They cannot be created via Terraform — use `terraform import` instead:
#
#   terraform import elasticstack_kibana_stream.existing_logs default/logs-myapp-default
#
resource "elasticstack_kibana_stream" "existing_logs" {
  name     = "logs-myapp-default"
  space_id = "default"

  classic_config = {
    # Add a processing step to enrich ingest documents
    processing_steps = [
      jsonencode({ action = "grok", from = "message", patterns = ["%%{TIMESTAMP_ISO8601:@timestamp} %%{LOGLEVEL:log.level} %%{GREEDYDATA:message}"] }),
    ]

    # Override field types for classic stream fields
    field_overrides_json = jsonencode({
      "host.name" = { type = "keyword" }
    })

    lifecycle_json = jsonencode({
      dsl = {
        data_retention = "14d"
      }
    })
  }
}
