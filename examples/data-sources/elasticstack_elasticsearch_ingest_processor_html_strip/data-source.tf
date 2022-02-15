provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_html_strip" "html_strip" {
  field = "foo"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "strip-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_html_strip.html_strip.json
  ]
}
