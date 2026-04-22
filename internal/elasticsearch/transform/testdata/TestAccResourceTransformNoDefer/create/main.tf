variable "transform_name" {
  type = string
}

variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name = var.index_name

  alias = [{
    name = "test_alias_1"
  }]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection    = false
  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name        = var.transform_name
  description = "test description"

  source {
    indices = [elasticstack_elasticsearch_index.test_index.name]
  }

  destination {
    index = "dest_index_for_transform"
  }

  pivot = jsonencode({
    "group_by" : {
      "customer_id" : {
        "terms" : {
          "field" : "customer_id",
          "missing_bucket" : true
        }
      }
    },
    "aggregations" : {
      "max_price" : {
        "max" : {
          "field" : "taxful_total_price"
        }
      }
    }
  })
  frequency = "5m"
  enabled   = false

  defer_validation = false
  timeout          = "1m"
}
