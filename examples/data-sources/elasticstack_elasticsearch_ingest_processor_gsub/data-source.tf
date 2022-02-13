provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_gsub" "gsub" {
  field       = "field1"
  pattern     = "\\."
  replacement = "-"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "gsub-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_gsub.gsub.json
  ]
}
