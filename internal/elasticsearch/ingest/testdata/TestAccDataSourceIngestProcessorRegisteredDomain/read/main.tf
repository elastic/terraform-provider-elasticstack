provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field        = "fqdn"
  target_field = "url"
}
