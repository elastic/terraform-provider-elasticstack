provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field          = "dns.question.name"
  target_field   = "domain_parts"
  description    = "Extract domain details from DNS question"
  if             = "ctx.dns?.question?.name != null"
  ignore_missing = true
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "registered domain lookup failed"
      }
    })
  ]
  tag = "registered-domain-update"
}
