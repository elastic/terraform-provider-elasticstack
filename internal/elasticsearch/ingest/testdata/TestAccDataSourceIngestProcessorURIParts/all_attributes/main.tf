provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uri_parts" "test" {
  field                = "request.uri"
  target_field         = "parsed_url"
  keep_original        = false
  remove_if_successful = true
  description          = "Parse URI parts from request"
  if                   = "ctx.request?.uri != null"
  ignore_failure       = true
  tag                  = "uri-parts-tag"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "uri parts failed"
      }
    })
  ]
}
