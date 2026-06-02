variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  number_of_replicas = 1
  refresh_interval   = "30s"

  analysis_analyzer = jsonencode({
    import_test = {
      type      = "custom"
      tokenizer = "standard"
      filter    = ["lowercase"]
    }
  })

  wait_for_active_shards = "1"
  master_timeout         = "30s"
  timeout                = "30s"

  deletion_protection = false
}
