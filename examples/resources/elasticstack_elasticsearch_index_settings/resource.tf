provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "my_index_settings" {
  index = "my-index"

  number_of_replicas = 2
  refresh_interval   = "30s"

  settings_json = jsonencode({
    max_result_window = 20000
  })
}
