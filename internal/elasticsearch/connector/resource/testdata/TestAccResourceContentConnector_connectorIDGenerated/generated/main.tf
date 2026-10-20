provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  service_type = "postgresql"
  name         = "TF acc generated id"
}
