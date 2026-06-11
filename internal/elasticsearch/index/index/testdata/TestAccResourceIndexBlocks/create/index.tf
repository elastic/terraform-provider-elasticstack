variable "index_name" {
  description = "The index name"
  type        = string
}

variable "blocks_write" {
  description = "Whether to block write operations"
  type        = bool
}

variable "blocks_read" {
  description = "Whether to block read operations"
  type        = bool
  default     = false
}

variable "blocks_metadata" {
  description = "Whether to block metadata reads and writes"
  type        = bool
  default     = false
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_blocks" {
  name                          = var.index_name
  blocks_write                  = var.blocks_write
  blocks_read                   = var.blocks_read
  blocks_read_only              = false
  blocks_read_only_allow_delete = false
  blocks_metadata               = var.blocks_metadata

  deletion_protection = false
}
