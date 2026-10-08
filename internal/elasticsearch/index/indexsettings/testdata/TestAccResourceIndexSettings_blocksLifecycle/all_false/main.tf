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

  blocks_read_only              = false
  blocks_read_only_allow_delete = false
  blocks_read                   = false
  blocks_write                  = false
  blocks_metadata               = false
}
