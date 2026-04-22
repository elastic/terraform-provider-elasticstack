provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = "validation"

  elasticsearch_connection {
    key_data = "pem-data"
  }
}
