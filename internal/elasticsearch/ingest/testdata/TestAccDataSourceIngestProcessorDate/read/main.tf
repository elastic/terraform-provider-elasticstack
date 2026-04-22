provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "test" {
  field        = "initial_date"
  target_field = "timestamp"
  formats      = ["dd/MM/yyyy HH:mm:ss"]
  timezone     = "Europe/Amsterdam"
}
