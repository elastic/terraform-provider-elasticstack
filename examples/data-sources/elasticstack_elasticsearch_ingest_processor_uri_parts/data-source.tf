provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uri_parts" "parts" {
  field                = "input_field"
  target_field         = "url"
  keep_original        = true
  remove_if_successful = false
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "parts-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_uri_parts.parts.json
  ]
}
