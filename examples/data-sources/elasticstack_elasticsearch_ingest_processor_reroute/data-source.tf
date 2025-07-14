provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "reroute" {
  dataset = ["my-dataset"]
  namespace = ["my-namespace"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "reroute-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_reroute.reroute.json
  ]
}
