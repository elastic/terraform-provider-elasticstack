provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index = "test-index"

  settings_json = jsonencode({
    "some.future.dynamic.setting" = "value"
  })
}
