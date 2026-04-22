provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field          = "fqdn"
  target_field   = "url_parts"
  description    = "Extract registered domain parts"
  if             = "ctx.fqdn != null"
  ignore_missing = true
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "registered domain failed"
      }
    })
  ]
  tag = "registered-domain"
}
