provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set_security_user" "user" {
  field      = "user"
  properties = ["username", "realm"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "user-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_set_security_user.user.json
  ]
}
