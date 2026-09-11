provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_rename" "test" {
  field        = "provider"
  target_field = "cloud.provider"
  description  = "Rename provider field with multiple on_failure handlers"
  if           = "ctx.provider == null"
  tag          = "rename-provider-multi"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "rename failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "rename_error"
      }
    })
  ]
}
