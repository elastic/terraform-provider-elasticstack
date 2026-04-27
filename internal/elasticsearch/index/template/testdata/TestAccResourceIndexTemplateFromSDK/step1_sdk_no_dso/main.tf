variable "template_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Same as step1_sdk but omits template.data_stream_options (Elasticsearch < 9.1.0).
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

    # Modern Elasticsearch echoes generic routing into index_routing/search_routing on read.
    # SDKv2 0.14.5 stores those echoes in state; omitting them here causes a non-empty refresh
    # plan (TypeSet replace) before we switch to the Plugin Framework implementation.
    # Match SDK readback on modern ES: echoed index/search routing without a separate top-level routing field.
    alias {
      name           = "routing_only_alias"
      index_routing  = "shard-a"
      search_routing = "shard-a"
      is_hidden      = false
      is_write_index = false
    }

    lifecycle {
      data_retention = "7d"
    }
  }
}
