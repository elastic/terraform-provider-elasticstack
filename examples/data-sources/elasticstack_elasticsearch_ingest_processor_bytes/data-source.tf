provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_bytes" "bytes" {
  field = "file.size"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "bytes-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_bytes.bytes.json
  ]
}
