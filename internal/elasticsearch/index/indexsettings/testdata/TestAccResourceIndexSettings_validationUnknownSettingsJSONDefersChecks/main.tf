provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "base" {
  name                = "tf-test-indexsettings-unknown-plan"
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

# `id` is computed, so the jsonencode expression (and with it settings_json)
# is unknown while the resource set is being planned.
resource "elasticstack_elasticsearch_index_settings" "depends" {
  index            = elasticstack_elasticsearch_index.base.name
  refresh_interval = "10s"
}

resource "elasticstack_elasticsearch_index_settings" "unknown_value" {
  index = elasticstack_elasticsearch_index.base.name

  settings_json = jsonencode({
    "index.refresh_interval" = elasticstack_elasticsearch_index_settings.depends.id
  })
}
