provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "json_proc" {
  field        = "string_source"
  target_field = "json_target"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "json-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_json.json_proc.json
  ]
}
