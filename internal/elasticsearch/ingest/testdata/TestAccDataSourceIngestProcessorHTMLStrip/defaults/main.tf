provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_html_strip" "test" {
  field = "content.html"
}
