provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field              = "message"
  value              = "plain-text"
  override           = false
  ignore_empty_value = true
  media_type         = "text/plain"
}
