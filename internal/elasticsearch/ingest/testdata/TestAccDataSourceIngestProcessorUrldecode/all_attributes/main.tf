provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_urldecode" "test" {
  field          = "source.url"
  target_field   = "decoded.url"
  ignore_missing = true
  ignore_failure = true
  description    = "Decode URL-encoded field"
  if             = "ctx.source?.url != null"
  tag            = "urldecode-tag"
}
