variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name               = var.index_name
  number_of_shards   = 1
  number_of_replicas = 0

  mappings = jsonencode({
    properties = {
      title  = { type = "text" }
      status = { type = "keyword" }
    }
  })

  analysis_filter = jsonencode({
    english_stop = {
      type      = "stop"
      stopwords = "_english_"
    }
  })

  analysis_analyzer = jsonencode({
    custom_english = {
      type      = "custom"
      tokenizer = "standard"
      filter    = ["lowercase", "english_stop"]
    }
  })

  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
