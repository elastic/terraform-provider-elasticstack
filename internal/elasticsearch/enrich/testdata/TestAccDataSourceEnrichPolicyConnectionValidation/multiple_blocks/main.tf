provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = "validation"

  elasticsearch_connection {
    endpoints = ["http://localhost:9200"]
  }

  elasticsearch_connection {
    endpoints = ["http://localhost:9200"]
  }
}
