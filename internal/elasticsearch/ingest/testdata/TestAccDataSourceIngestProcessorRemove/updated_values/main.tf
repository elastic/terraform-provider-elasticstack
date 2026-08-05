provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field = [
    "host.name",
    "user.name",
  ]
  description = "Remove host and user fields"
  if          = "ctx.host != null"
  tag         = "remove-host-user-fields"
}
