provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "reroute" {
  destination = "logs-generic-default"
  dataset     = "generic"
  namespace   = "default"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "reroute-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_reroute.reroute.json
  ]
}