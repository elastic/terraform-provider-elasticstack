provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uri_parts" "test" {
  field                = "input_field"
  target_field         = "url"
  keep_original        = true
  remove_if_successful = false
}
