provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_html_strip" "test" {
  field          = "body.html"
  target_field   = "body.plain"
  ignore_missing = true
  description    = "Strip HTML markup from body content"
  if             = "ctx.body?.html != null"
  ignore_failure = true
  tag            = "html-strip-tag"
}
