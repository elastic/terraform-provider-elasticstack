provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_user_agent" "test" {
  field = "agent"
}
