provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_geoip" "geoip" {
  field = "ip"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "geoip-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_geoip.geoip.json
  ]
}
