provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index = "test-index"

  settings_json = jsonencode({
    refresh_interval  = null
    max_result_window = 20000
  })
}
