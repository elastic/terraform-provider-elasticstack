provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_urldecode" "urldecode" {
  field = "my_url_to_decode"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "urldecode-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_urldecode.urldecode.json
  ]
}
