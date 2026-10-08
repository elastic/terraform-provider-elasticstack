variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index = elasticstack_elasticsearch_index.test.name

  settings_json = jsonencode({
    "index.refresh_interval" = "10s"
  })
}
