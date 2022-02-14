data "elasticstack_elasticsearch_ingest_processor_set" "set_count" {
  field = "count"
  value = 1
}

data "elasticstack_elasticsearch_ingest_processor_json" "parse_string_source" {
  field        = "string_source"
  target_field = "json_target"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "ingest" {
  name = "set-parse"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_set.set_count.json,
    data.elasticstack_elasticsearch_ingest_processor_json.parse_string_source.json
  ]
}
