variable "template_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "upgrade" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
  priority       = 100

  data_stream {
    hidden               = true
    allow_custom_routing = true
  }

  template {
    mappings = jsonencode({
      properties = { from_sdk = { type = "keyword" } }
    })
    settings = jsonencode({
      index = { number_of_shards = 1 }
    })

    alias {
      name    = "routing_only_alias"
      routing = "shard-a"
    }

    lifecycle {
      data_retention = "7d"
    }

    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "30d"
        }
      }
    }
  }
}
