provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "domain" {
  field        = "fqdn"
  target_field = "url"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "domain-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_registered_domain.domain.json
  ]
}
