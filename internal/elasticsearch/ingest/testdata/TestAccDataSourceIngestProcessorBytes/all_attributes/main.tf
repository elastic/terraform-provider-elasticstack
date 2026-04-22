provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_bytes" "test" {
  field          = "document.size"
  target_field   = "document.size_bytes"
  ignore_missing = true
  ignore_failure = true
  description    = "Convert document size to bytes"
  if             = "ctx.document?.size != null"
  tag            = "bytes-tag"
}
