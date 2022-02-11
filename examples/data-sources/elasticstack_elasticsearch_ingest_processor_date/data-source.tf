provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "date" {
  field        = "initial_date"
  target_field = "timestamp"
  formats      = ["dd/MM/yyyy HH:mm:ss"]
  timezone     = "Europe/Amsterdam"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "date-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_date.date.json
  ]
}
