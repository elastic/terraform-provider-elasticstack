provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  description    = "converts the content of the id field to an integer"
  field          = "id"
  target_field   = "converted_id"
  type           = "integer"
  if             = "ctx.id != null"
  ignore_missing = true
  ignore_failure = true
  tag            = "convert-tag"
}
