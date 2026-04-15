provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field = "host.name"
}
