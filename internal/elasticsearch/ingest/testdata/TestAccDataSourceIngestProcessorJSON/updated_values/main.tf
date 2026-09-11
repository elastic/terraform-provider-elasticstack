provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field        = "updated_string_source"
  target_field = "updated_json_target"
  description  = "Parse updated document JSON"
  if           = "ctx.updated_document?.json != null"
  tag          = "updated-json-tag"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "updated json processor failed"
      }
    })
  ]
}
