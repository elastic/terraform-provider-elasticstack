provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_rename" "test" {
  field          = "provider"
  target_field   = "cloud.provider"
  description    = "Rename provider field"
  if             = "ctx.provider != null"
  ignore_missing = true
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "rename failed"
      }
    })
  ]
  tag = "rename-provider"
}
