provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_urldecode" "test" {
  field = "my_url_to_decode"
}
