provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test" {
  iana_number = 6
}
