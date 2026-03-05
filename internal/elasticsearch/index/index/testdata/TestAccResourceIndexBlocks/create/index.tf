variable "index_name" {
  description = "The index name"
  type        = string
}

variable "blocks_write" {
  description = "Whether to block write operations"
  type        = bool
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_blocks" {
  name             = var.index_name
  blocks_write     = var.blocks_write
  blocks_read      = false
  blocks_read_only = false
  blocks_metadata  = false

  deletion_protection = false
}
