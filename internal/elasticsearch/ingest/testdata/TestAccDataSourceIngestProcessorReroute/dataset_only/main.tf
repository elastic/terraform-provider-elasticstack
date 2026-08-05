provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test" {
  dataset     = "metrics"
  description = "Route metrics by dataset"
}
