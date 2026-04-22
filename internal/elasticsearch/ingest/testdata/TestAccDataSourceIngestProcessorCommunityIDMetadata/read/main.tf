provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test" {
  description = "Compute the community ID"
  if          = "ctx.network != null"
  tag         = "community-id-tag"
}
