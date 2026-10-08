provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index = "test-index"

  settings_json = jsonencode({
    "index.query.default_field" = ["title", null]
  })
}
