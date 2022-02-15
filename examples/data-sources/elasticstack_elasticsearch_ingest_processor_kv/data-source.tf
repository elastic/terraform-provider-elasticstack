provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "kv" {
  field       = "message"
  field_split = " "
  value_split = "="

  exclude_keys = ["tags"]
  prefix       = "setting_"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "kv-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_kv.kv.json
  ]
}
