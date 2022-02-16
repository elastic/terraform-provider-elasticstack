provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "fingerprint" {
  fields = ["user"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "fingerprint-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_fingerprint.fingerprint.json
  ]
}
