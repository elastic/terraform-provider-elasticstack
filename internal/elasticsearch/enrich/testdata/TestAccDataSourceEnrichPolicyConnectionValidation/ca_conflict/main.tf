provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = "validation"

  elasticsearch_connection {
    ca_file = "/tmp/ca.pem"
    ca_data = "pem-data"
  }
}
