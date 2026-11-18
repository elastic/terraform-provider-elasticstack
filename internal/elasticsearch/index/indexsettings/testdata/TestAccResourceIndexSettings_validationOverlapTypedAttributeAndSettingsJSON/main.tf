provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index              = "test-index"
  number_of_replicas = 1

  settings_json = jsonencode({
    number_of_replicas = 2
  })
}
