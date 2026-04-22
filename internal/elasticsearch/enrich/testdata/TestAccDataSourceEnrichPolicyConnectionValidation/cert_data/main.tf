provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = "validation"

  elasticsearch_connection {
    cert_data = "pem-data"
  }
}
