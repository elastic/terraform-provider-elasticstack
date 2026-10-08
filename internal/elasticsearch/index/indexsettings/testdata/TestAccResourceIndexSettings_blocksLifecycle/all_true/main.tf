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

  blocks_read_only              = true
  blocks_read_only_allow_delete = true
  blocks_read                   = true
  blocks_write                  = true
  blocks_metadata               = false
}
