variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  deletion_protection = false

  alias = [
    {
      name           = "${var.index_name}-hidden"
      is_hidden      = true
      is_write_index = false
    },
    {
      name           = "${var.index_name}-write"
      is_write_index = true
    },
  ]

  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}
